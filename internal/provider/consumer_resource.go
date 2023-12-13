package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"terraform-provider-nats/internal/nats"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &consumerResource{}
var _ resource.ResourceWithImportState = &consumerResource{}

func NewConsumerResource() resource.Resource {
	return &consumerResource{}
}

type consumerResource struct {
	client nats.Client
}

type consumerResourceModel struct {
	StreamName types.String `tfsdk:"stream_name"`
	Name       types.String `tfsdk:"name"`

	Mode           types.String   `tfsdk:"mode"`
	DeliverPolicy  types.String   `tfsdk:"deliver_policy"`
	AckPolicy      types.String   `tfsdk:"ack_policy"`
	FilterSubjects []types.String `tfsdk:"filter_subjects"`

	// Push-Specific
	DeliverSubject types.String `tfsdk:"deliver_subject"`
	DeliverGroup   types.String `tfsdk:"deliver_group"`
}

func (r *consumerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumer"
}

func (r *consumerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Consumer resource",
		Attributes: map[string]schema.Attribute{
			"stream_name": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"mode": schema.StringAttribute{
				Description: "The consumer mode. Possible values: push, pull.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("push", "pull")},
			},
			"deliver_policy": schema.StringAttribute{
				Description: "The point in the stream to receive messages from. Possible values: all (default), new, last.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
				Validators:  []validator.String{stringvalidator.OneOf("all", "new", "last")},
			},
			"ack_policy": schema.StringAttribute{
				Description: "The requirement of client acknowledgements. Possible values: none (default), all, explicit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Validators:  []validator.String{stringvalidator.OneOf("none", "all", "explicit")},
			},
			"filter_subjects": schema.ListAttribute{
				Description: "A set of subjects that overlap with the subjects bound to the stream to filter delivery to subscribers. Default is all stream subjects (no filtering).",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, nil)),
			},
			// Push-specific
			"deliver_subject": schema.StringAttribute{
				Description: "The subject to deliver messages to. The server will push messages to client subscribed to this subject. Must be set if mode = push.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"deliver_group": schema.StringAttribute{
				Description: "The queue group name which, if specified, is then used to distribute the messages between the subscribers to the consumer. Used only if mode = push",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *consumerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(nats.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected nats.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *consumerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 1. Read plan
	var data consumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Create the resource
	consumerConfig, err := toConsumerConfig(data)
	if err != nil {
		resp.Diagnostics.AddError("Validation error", err.Error())
		return
	}
	consumerInfo, err := r.client.CreateConsumer(data.StreamName.ValueString(), consumerConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to create consumer: %s", err))
		return
	}
	// 3. Write state
	data = fromConsumerInfo(consumerInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *consumerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 1. Read current state
	var data consumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Get the resource
	consumerInfo, err := r.client.GetConsumer(data.StreamName.ValueString(), data.Name.ValueString())
	if err != nil {
		if errors.Is(err, nats.ErrNotFound) {
			resp.Diagnostics.AddWarning("Resource not found", "couldn't find the consumer, possibly deleted outside terraform")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to read consumer: %s", err))
		return
	}
	// 3. Write new state
	data = fromConsumerInfo(consumerInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *consumerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 1. Read plan & current state
	var plan consumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state consumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Validate id is not changed
	if plan.Name != state.Name || plan.StreamName != state.StreamName {
		resp.Diagnostics.AddError(
			"Cannot change consumer or stream name",
			"The consumer is identified by its name + stream name. If you wish to change the name, you must create a new consumer.",
		)
		return
	}
	// 3. Update resource
	consumerConfig, err := toConsumerConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("Validation error", err.Error())
		return
	}
	consumerInfo, err := r.client.UpdateConsumer(plan.StreamName.ValueString(), consumerConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to update consumer: %s", err))
		return
	}
	// 4. Write new state
	state = fromConsumerInfo(consumerInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *consumerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 1. Read current state
	var state consumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Delete the resource
	err := r.client.DeleteConsumer(state.StreamName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to delete consumer: %s", err))
		return
	}
}

func (r *consumerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp) // TODO(sagheer): id is name + stream_name

	tokens := strings.Split(req.ID, "#")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		resp.Diagnostics.AddError("Invalid import id", "The import id must be of the format 'stream_name#consumer_name'")
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stream_name"), tokens[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), tokens[1])...)
}

func toConsumerConfig(data consumerResourceModel) (nats.ConsumerConfig, error) {
	if err := validateMode(data); err != nil {
		return nats.ConsumerConfig{}, err
	}
	return nats.ConsumerConfig{
		Name:           data.Name.ValueString(),
		Durable:        data.Name.ValueString(),
		DeliverPolicy:  nats.ToDeliverPolicy(data.DeliverPolicy.ValueString()),
		AckPolicy:      nats.ToAckPolicy(data.AckPolicy.ValueString()),
		FilterSubjects: convertSlice(data.FilterSubjects, (types.String).ValueString),
		DeliverSubject: data.DeliverSubject.ValueString(),
		DeliverGroup:   data.DeliverGroup.ValueString(),
	}, nil
}

func fromConsumerInfo(consumerInfo nats.ConsumerInfo) consumerResourceModel {
	var mode = "pull"
	if consumerInfo.Config.DeliverSubject != "" {
		mode = "push"
	}
	return consumerResourceModel{
		StreamName:     types.StringValue(consumerInfo.Stream),
		Name:           types.StringValue(consumerInfo.Name),
		Mode:           types.StringValue(mode),
		DeliverPolicy:  types.StringValue(nats.FromDeliverPolicy(consumerInfo.Config.DeliverPolicy)),
		AckPolicy:      types.StringValue(nats.FromAckPolicy(consumerInfo.Config.AckPolicy)),
		FilterSubjects: convertSlice(consumerInfo.Config.FilterSubjects, types.StringValue),
		DeliverSubject: types.StringValue(consumerInfo.Config.DeliverSubject),
		DeliverGroup:   types.StringValue(consumerInfo.Config.DeliverGroup),
	}
}

func validateMode(data consumerResourceModel) error {
	if data.Mode.ValueString() == "pull" && data.DeliverSubject.ValueString() != "" {
		return fmt.Errorf("Attribute 'deliver_subject' must not be set if 'mode' is 'pull'")
	}
	if data.Mode.ValueString() == "push" && data.DeliverSubject.ValueString() == "" {
		return fmt.Errorf("Attribute 'deliver_subject' must be set if 'mode' is 'push'")
	}
	return nil
}
