<script lang="ts">
	import {
		type Project,
		type Memory,
		getMemories,
		deleteAllMemories,
		deleteMemory,
		updateMemory
	} from '$lib/services';
	import { X, Trash2, RefreshCcw, Edit, Check, X as XIcon } from 'lucide-svelte/icons';
	import { fade } from 'svelte/transition';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import errors from '$lib/stores/errors.svelte';
	import Confirm from './Confirm.svelte';

	interface Props {
		project?: Project;
	}

	let { project = $bindable() }: Props = $props();
	let dialog = $state<HTMLDialogElement>();
	let memories = $state<Memory[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let toDeleteAll = $state(false);
	let editingMemoryId = $state<string | null>(null);
	let editContent = $state('');

	export function show(projectToUse?: Project) {
		if (projectToUse) {
			project = projectToUse;
		}

		dialog?.showModal();
		loadMemories();
	}

	function closeDialog() {
		dialog?.close();
	}

	async function loadMemories() {
		if (!project) return;

		loading = true;
		error = null;
		try {
			const result = await getMemories(project.assistantID, project.id);
			memories = result.items || [];
		} catch (err) {
			// Ignore 404 errors (memory tool not configured or no memories)
			if (err instanceof Error && err.message.includes('404')) {
				memories = [];
			} else {
				// For all other errors, append to errors store
				errors.append(err);
				error = 'Failed to load memories';
			}
		} finally {
			loading = false;
		}
	}

	async function deleteAll() {
		if (!project) return;

		loading = true;
		error = null;
		try {
			await deleteAllMemories(project.assistantID, project.id);
			memories = [];
		} catch (err) {
			errors.append(err);
			error = 'Failed to delete all memories';
		} finally {
			loading = false;
			toDeleteAll = false;
		}
	}

	async function deleteOne(memoryId: string) {
		if (!project) return;

		loading = true;
		error = null;
		try {
			await deleteMemory(project.assistantID, project.id, memoryId);
			memories = memories.filter((memory) => memory.id !== memoryId);
		} catch (err) {
			errors.append(err);
			error = 'Failed to delete memory';
		} finally {
			loading = false;
		}
	}

	function startEdit(memory: Memory) {
		editingMemoryId = memory.id;
		editContent = memory.content;
	}

	function cancelEdit() {
		editingMemoryId = null;
		editContent = '';
	}

	async function saveEdit() {
		if (!project || !editingMemoryId) return;

		loading = true;
		error = null;
		try {
			const updatedMemory = await updateMemory(
				project.assistantID,
				project.id,
				editingMemoryId,
				editContent
			);
			// Update the memory in the list
			memories = memories.map((memory) => (memory.id === editingMemoryId ? updatedMemory : memory));
			editingMemoryId = null;
			editContent = '';
		} catch (err) {
			errors.append(err);
			error = 'Failed to update memory';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateString: string): string {
		if (!dateString) return '';

		try {
			const date = new Date(dateString);
			return date.toLocaleString();
		} catch (_e) {
			return '';
		}
	}
</script>

<dialog
	bind:this={dialog}
	class="bg-surface1 border-surface3 max-h-[90vh] min-h-[300px] w-2/3 max-w-[900px] min-w-[600px] overflow-visible rounded-lg border p-5"
>
	<div class="flex h-full max-h-[calc(90vh-40px)] flex-col">
		<button class="absolute top-0 right-0 p-3" onclick={closeDialog}>
			<X class="icon-default" />
		</button>
		<h1 class="text-text1 mb-4 text-xl font-semibold">Memories</h1>

		{#if error}
			<div class="mb-4 rounded bg-red-100 p-3 text-red-800">{error}</div>
		{/if}

		<div class="mb-4 flex items-center justify-between">
			<span class="text-text2 text-sm">{memories.length} memories</span>
			<div class="flex gap-2">
				<button class="icon-button" onclick={() => loadMemories()} use:tooltip={'Refresh Memories'}>
					<RefreshCcw class="size-4" />
				</button>
				<button
					class="button-small bg-red-500 hover:bg-red-600 disabled:cursor-not-allowed disabled:opacity-50"
					onclick={() => (toDeleteAll = true)}
					disabled={loading || memories.length === 0}
				>
					<Trash2 class="size-4" />
					Delete All
				</button>
			</div>
		</div>

		<div class="min-h-0 flex-1 overflow-auto">
			{#if loading}
				<div in:fade class="flex justify-center py-10">
					<div
						class="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent"
					></div>
				</div>
			{:else if memories.length === 0}
				<p in:fade class="text-gray pt-6 pb-3 text-center text-sm dark:text-gray-300">
					No memories stored
				</p>
			{:else}
				<div class="overflow-auto">
					<table class="w-full text-left">
						<thead class="bg-surface1 sticky top-0 z-10">
							<tr class="border-surface3 border-b">
								<th class="text-text1 py-2 text-sm font-medium whitespace-nowrap">Created</th>
								<th class="text-text1 w-full py-2 text-sm font-medium">Content</th>
								<th class="text-text1 py-2 text-sm font-medium"></th>
							</tr>
						</thead>
						<tbody>
							{#each memories as memory (memory.id)}
								<tr class="border-surface3 group hover:bg-surface2 border-b">
									<td class="text-text2 py-3 pr-4 text-xs whitespace-nowrap"
										>{formatDate(memory.createdAt)}</td
									>
									<td
										class="text-text1 max-w-[450px] py-3 pr-4 text-sm break-words break-all hyphens-auto"
									>
										{#if editingMemoryId === memory.id}
											<textarea
												bind:value={editContent}
												class="border-surface3 bg-surface2 text-text1 min-h-[80px] w-full rounded border p-2"
												rows="3"
											></textarea>
										{:else}
											{memory.content}
										{/if}
									</td>
									<td class="py-3 whitespace-nowrap">
										{#if editingMemoryId === memory.id}
											<div class="flex gap-2">
												<button
													class="icon-button text-green-500"
													onclick={saveEdit}
													use:tooltip={'Save changes'}
												>
													<Check class="h-4 w-4" />
												</button>
												<button
													class="icon-button text-red-500"
													onclick={cancelEdit}
													use:tooltip={'Cancel'}
												>
													<XIcon class="h-4 w-4" />
												</button>
											</div>
										{:else}
											<div class="flex gap-2">
												<button
													class="icon-button"
													onclick={() => startEdit(memory)}
													disabled={loading}
													use:tooltip={'Edit memory'}
												>
													<Edit class="h-4 w-4" />
												</button>
												<button
													class="icon-button"
													onclick={() => deleteOne(memory.id)}
													disabled={loading}
													use:tooltip={'Delete memory'}
												>
													<Trash2 class="h-4 w-4" />
												</button>
											</div>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</dialog>

<Confirm
	msg={'Are you sure you want to delete all memories?'}
	show={toDeleteAll}
	onsuccess={deleteAll}
	oncancel={() => (toDeleteAll = false)}
/>
