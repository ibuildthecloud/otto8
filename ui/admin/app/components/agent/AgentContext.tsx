import {
	ReactNode,
	createContext,
	useCallback,
	useContext,
	useState,
} from "react";
import useSWR, { mutate } from "swr";

import { Agent } from "~/lib/model/agents";
import { AgentService } from "~/lib/service/api/agentService";

import { useAsync } from "~/hooks/useAsync";

interface AgentContextType {
	agent: Agent;
	agentId: string;
	updateAgent: (agent: Agent) => Promise<unknown>;
	refreshAgent: (agent?: Agent) => Promise<unknown>;
	isUpdating: boolean;
	error?: unknown;
	lastUpdated?: Date;
}

const AgentContext = createContext<AgentContextType | undefined>(undefined);

export function AgentProvider({
	children,
	agent,
}: {
	children: ReactNode;
	agent: Agent;
}) {
	const agentId = agent.id;

	const getAgent = useSWR(
		AgentService.getAgentById.key(agentId),
		({ agentId }) => AgentService.getAgentById(agentId),
		{ fallbackData: agent }
	);

	const [lastUpdated, setLastSaved] = useState<Date>();

	const handleUpdateAgent = useCallback(
		(updatedAgent: Agent) =>
			AgentService.updateAgent({ id: agentId, agent: updatedAgent })
				.then((updatedAgent) => {
					getAgent.mutate(updatedAgent);
					mutate(AgentService.getAgents.key());
					setLastSaved(new Date());
				})
				.catch(console.error),
		[agentId, getAgent]
	);

	const updateAgent = useAsync(handleUpdateAgent);

	const refreshAgent = getAgent.mutate;

	return (
		<AgentContext.Provider
			value={{
				agentId,
				agent: getAgent.data ?? agent,
				updateAgent: updateAgent.executeAsync,
				refreshAgent,
				isUpdating: updateAgent.isLoading,
				lastUpdated,
				error: updateAgent.error,
			}}
		>
			{children}
		</AgentContext.Provider>
	);
}

export function useAgent() {
	const context = useContext(AgentContext);
	if (context === undefined) {
		throw new Error("useChat must be used within a ChatProvider");
	}
	return context;
}
