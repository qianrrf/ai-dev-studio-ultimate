import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { ProjectCodexControl } from "./ProjectCodexControl";

const h = vi.hoisted(() => ({ get: vi.fn(), put: vi.fn(), post: vi.fn() }));
vi.mock("../lib/api-client", () => ({ apiClient: { GET: h.get, PUT: h.put, POST: h.post }, apiErrorMessage: (e: { message?: string }) => e?.message ?? "请求失败" }));
afterEach(cleanup);
beforeEach(() => {
	vi.clearAllMocks();
	h.get.mockImplementation((path: string) => Promise.resolve({ data: path.endsWith("control") ? { projectId: "demo", ownerKind: "external", ownerId: "old", revision: 7 } : { requests: [{ id: "brief-1", brief: "修复计算器" }] } }));
	h.put.mockResolvedValue({ data: { ownerKind: "external", ownerId: "external-codex", revision: 8 } });
});
function show() {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
	render(<QueryClientProvider client={client}><ProjectCodexControl projectId="demo" /></QueryClientProvider>);
}
it("loads retained requirements without launching a model and uses the observed revision on takeover", async () => {
	show();
	await screen.findByText("修复计算器");
	expect(h.post).not.toHaveBeenCalled();
	fireEvent.click(screen.getByRole("button", { name: "Use external Codex" }));
	await waitFor(() => expect(h.put).toHaveBeenCalledWith("/api/v1/projects/{id}/control", { params: { path: { id: "demo" } }, body: { ownerKind: "external", ownerId: "external-codex", expectedRevision: 7 } }));
});
it("keeps a takeover conflict visible and does not silently retry or launch another controller", async () => {
	h.put.mockResolvedValue({ error: { message: "主控已经切换" } });
	show();
	await screen.findByText(/版本 7/);
	fireEvent.click(screen.getByRole("button", { name: "释放主控" }));
	await screen.findByRole("alert");
	expect(screen.getByRole("alert")).toHaveTextContent("主控已经切换");
	expect(h.put).toHaveBeenCalledTimes(1);
	expect(h.post).not.toHaveBeenCalled();
});
