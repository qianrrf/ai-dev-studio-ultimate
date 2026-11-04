package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// Mirrored wire types keep the CLI independent from HTTP controller packages.
type controlClaim struct {
	OwnerID  string `json:"ownerId"`
	Revision int64  `json:"revision"`
}

type projectControlRequest struct {
	OwnerKind        string `json:"ownerKind"`
	OwnerID          string `json:"ownerId"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

func newProjectControlCommand(ctx *commandContext) *cobra.Command {
	var owner, entry string
	var revision int64
	var release bool
	cmd := &cobra.Command{Use: "control <project>", Short: "Query, take over, or release project controller ownership", Args: exactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		route := "projects/" + url.PathEscape(args[0]) + "/control"
		var result json.RawMessage
		if !release && owner == "" {
			if cmd.Flags().Changed("revision") || cmd.Flags().Changed("entry") {
				return usageError{fmt.Errorf("--revision and --entry require --owner or --release")}
			}
			if err := ctx.getJSON(cmd.Context(), route, &result); err != nil {
				return err
			}
		} else {
			if !cmd.Flags().Changed("revision") || revision < 0 || (release && owner != "") {
				return usageError{fmt.Errorf("takeover/release requires --revision >= 0; --owner and --release are exclusive")}
			}
			if entry != "internal" && entry != "external" {
				return usageError{fmt.Errorf("--entry must be internal or external")}
			}
			if release {
				entry = "unowned"
			}
			if err := ctx.putJSON(cmd.Context(), route, projectControlRequest{OwnerKind: entry, OwnerID: owner, ExpectedRevision: revision}, &result); err != nil {
				return err
			}
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}}
	cmd.Flags().StringVar(&owner, "owner", "", "Controller identity; internal uses the Codex orchestrator session ID")
	cmd.Flags().StringVar(&entry, "entry", "external", "Controller entry: internal or external")
	cmd.Flags().Int64Var(&revision, "revision", 0, "Expected current revision; 0 only for initial ownership")
	cmd.Flags().BoolVar(&release, "release", false, "Release ownership without deleting its revision history")
	return cmd
}

func newProjectRequestCommand(ctx *commandContext) *cobra.Command {
	var id string
	cmd := &cobra.Command{Use: "request <project> <brief>", Short: "Submit a durable requirement to the Codex controller inbox", Args: exactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(id) == "" || strings.TrimSpace(args[1]) == "" {
			return usageError{fmt.Errorf("--id and a non-empty brief are required")}
		}
		var result json.RawMessage
		in := struct {
			ID    string `json:"id"`
			Brief string `json:"brief"`
		}{id, args[1]}
		if err := ctx.postJSON(cmd.Context(), "projects/"+url.PathEscape(args[0])+"/requests", in, &result); err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}}
	cmd.Flags().StringVar(&id, "id", "", "Stable unique request ID; reuse it only to retry the same requirement")
	return cmd
}

func newProjectInboxCommand(ctx *commandContext) *cobra.Command {
	return &cobra.Command{Use: "inbox <project>", Short: "Read requirements preserved across controller handovers", Args: exactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		var result json.RawMessage
		if err := ctx.getJSON(cmd.Context(), "projects/"+url.PathEscape(args[0])+"/requests", &result); err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), result)
	}}
}
