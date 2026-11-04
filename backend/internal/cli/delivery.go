package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// Wire types are deliberately mirrored; CLI only talks to daemon HTTP.
type deliveryRequest struct {
	ID       string       `json:"id"`
	Control  controlClaim `json:"control"`
	Kind     string       `json:"kind"`
	BaseSHA  string       `json:"baseSha"`
	HeadSHA  string       `json:"headSha"`
	TestID   string       `json:"testId,omitempty"`
	ReviewID string       `json:"reviewId,omitempty"`
	Command  []string     `json:"command,omitempty"`
	Message  string       `json:"message,omitempty"`
}

func newSessionDeliveryCommand(ctx *commandContext) *cobra.Command {
	var in deliveryRequest
	cmd := &cobra.Command{Use: "delivery <session> [-- executable args...]", Short: "Read delivery evidence or register a controller test/review/decision", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		route := "sessions/" + url.PathEscape(args[0]) + "/delivery"
		var result json.RawMessage
		if in.Kind == "" {
			if len(args) != 1 || cmd.Flags().Changed("request-id") || cmd.Flags().Changed("controller") || cmd.Flags().Changed("control-revision") || cmd.Flags().Changed("base") || cmd.Flags().Changed("head") || cmd.Flags().Changed("test-id") || cmd.Flags().Changed("review-id") || cmd.Flags().Changed("message") {
				return usageError{fmt.Errorf("delivery options require --kind")}
			}
			if err := ctx.getJSON(cmd.Context(), route, &result); err != nil {
				return err
			}
		} else {
			if in.ID == "" || in.Control.OwnerID == "" || in.Control.Revision <= 0 || in.BaseSHA == "" || in.HeadSHA == "" {
				return usageError{fmt.Errorf("--request-id, --controller, --control-revision, --base and --head are required")}
			}
			switch in.Kind {
			case "test", "review", "verdict", "rework", "accept":
			default:
				return usageError{fmt.Errorf("invalid --kind")}
			}
			in.Command = args[1:]
			if (in.Kind == "test") != (len(in.Command) > 0) {
				return usageError{fmt.Errorf("only test requires a command after --")}
			}
			if err := ctx.postJSON(cmd.Context(), route, in, &result); err != nil {
				return err
			}
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}}
	cmd.Flags().StringVar(&in.Kind, "kind", "", "test, review, verdict, rework or accept; omitted reads evidence")
	cmd.Flags().StringVar(&in.ID, "request-id", "", "Stable operation ID; incomplete receipts must be reconciled")
	cmd.Flags().StringVar(&in.Control.OwnerID, "controller", "", "Current controller identity")
	cmd.Flags().Int64Var(&in.Control.Revision, "control-revision", 0, "Current controller revision")
	cmd.Flags().StringVar(&in.BaseSHA, "base", "", "Full base commit SHA")
	cmd.Flags().StringVar(&in.HeadSHA, "head", "", "Full committed worker head SHA")
	cmd.Flags().StringVar(&in.TestID, "test-id", "", "Successful matching test evidence ID")
	cmd.Flags().StringVar(&in.ReviewID, "review-id", "", "Review operation ID for verdict; verdict operation ID for rework/accept")
	cmd.Flags().StringVar(&in.Message, "message", "", "Review brief, rework instructions or acceptance conclusion")
	return cmd
}
