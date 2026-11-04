import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { apiClient, apiErrorMessage } from "../lib/api-client";
import { Button } from "./ui/button";

// Ownership and inbox facts live in the daemon. Opening settings never starts a model.
export function ProjectCodexControl({ projectId }: { projectId: string }) {
	const queryClient = useQueryClient();
	const [owner, setOwner] = useState("external-codex");
	const key = ["project-codex-control", projectId];
	const params = { path: { id: projectId } };
	const control = useQuery({
		queryKey: key,
		refetchInterval: 5000,
		queryFn: async () => {
			const { data, error } = await apiClient.GET("/api/v1/projects/{id}/control", { params });
			if (error || !data) throw new Error(apiErrorMessage(error));
			return data;
		},
	});
	const inbox = useQuery({
		queryKey: ["project-codex-inbox", projectId],
		refetchInterval: 5000,
		queryFn: async () => {
			const { data, error } = await apiClient.GET("/api/v1/projects/{id}/requests", { params });
			if (error || !data) throw new Error(apiErrorMessage(error));
			return data.requests;
		},
	});
	const change = useMutation({
		mutationFn: async (entry: "internal" | "external" | "unowned") => {
			if (!control.data) throw new Error("主控状态尚未加载");
			let revision = control.data.revision;
			let ownerId = entry === "unowned" ? "" : owner.trim();
			if (entry === "internal") {
				// First enable the policy so the existing spawn selects Codex Chat.
				if (revision === 0) {
					const { data, error } = await apiClient.PUT("/api/v1/projects/{id}/control", { params, body: { ownerKind: "unowned", ownerId: "", expectedRevision: 0 } });
					if (error || !data) throw new Error(apiErrorMessage(error));
					revision = data.revision;
				}
				const { data, error } = await apiClient.POST("/api/v1/sessions", {
					body: { projectId, kind: "orchestrator", harness: "codex", mode: "chat", approvalMode: "default" },
				});
				if (error || !data?.session) throw new Error(apiErrorMessage(error));
				ownerId = data.session.id;
			}
			const { data, error } = await apiClient.PUT("/api/v1/projects/{id}/control", { params, body: { ownerKind: entry, ownerId, expectedRevision: revision } });
			if (error || !data) throw new Error(apiErrorMessage(error));
			return data;
		},
		onSettled: async () => { await queryClient.invalidateQueries({ queryKey: key }); },
	});
	return <section className="flex flex-col gap-3" aria-label="Codex 主控">
		<h3 className="text-sm font-medium">Codex 主控</h3>
		<p className="text-sm text-settings-muted">新需求由当前主控安排；最多同时执行两个 DSH Worker。切换主控会保留已有任务。</p>
		{control.data && <p role="status" className="text-sm">{control.data.ownerKind === "unowned" ? "等待接管" : control.data.ownerKind === "internal" ? "产品内 Codex" : "外部 Codex"} · {control.data.ownerId || "未指定"} · 版本 {control.data.revision}</p>}
		<label className="text-sm">外部主控标识<input className="settings-input ml-2" value={owner} onChange={(e) => setOwner(e.target.value)} maxLength={128} /></label>
		<div className="flex flex-wrap gap-2">
			<Button type="button" disabled={!control.data || change.isPending} onClick={() => change.mutate("internal")}>产品内 Codex 接管</Button>
			<Button type="button" disabled={!control.data || change.isPending || !owner.trim()} onClick={() => change.mutate("external")}>外部 Codex 接管</Button>
			<Button type="button" disabled={!control.data || change.isPending || control.data.ownerKind === "unowned"} onClick={() => change.mutate("unowned")}>释放主控</Button>
		</div>
		{(change.error || control.error || inbox.error) && <p role="alert" className="text-sm text-error">{(change.error || control.error || inbox.error)?.message}</p>}
		<details><summary className="text-sm">需求记录（{inbox.data?.length ?? 0}）</summary>
			<ul className="flex flex-col gap-2">{inbox.data?.map((request) => <li key={request.id} className="text-sm"><span className="text-settings-muted">{request.id}</span><p className="whitespace-pre-wrap">{request.brief}</p></li>)}</ul>
		</details>
	</section>;
}
