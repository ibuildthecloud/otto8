import { ApiRoutes } from "~/lib/routers/apiRoutes";
import { request } from "~/lib/service/api/primitives";

async function invokeAsync({
	slug,
	prompt,
	thread,
}: {
	slug: string;
	prompt?: Nullish<string>;
	thread?: Nullish<string>;
}) {
	console.log("invokeAsync", slug, prompt, thread);

	const { data } = await request<{ threadID: string }>({
		url: ApiRoutes.invoke(slug, thread, { async: true }).url,
		method: "POST",
		data: prompt,
		errorMessage: "Failed to invoke agent",
	});

	return data;
}

export const InvokeService = {
	invokeAgent: invokeAsync,
	invokeWorkflow: invokeAsync,
};
