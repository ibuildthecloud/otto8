import { FEATURED_PROJECT_ORDER } from './constants';
import type { ProjectShare } from './services';

export const sortByFeaturedNameOrder = (a: ProjectShare, b: ProjectShare) => {
	const aName = (a.name?.toLowerCase() ?? '').trim();
	const bName = (b.name?.toLowerCase() ?? '').trim();
	const aIndex = FEATURED_PROJECT_ORDER.indexOf(aName);
	const bIndex = FEATURED_PROJECT_ORDER.indexOf(bName);
	if (aIndex === -1) return 1;
	if (bIndex === -1) return -1;
	return aIndex - bIndex;
};

export const sortShownToolsPriority = (a: string, b: string) => {
	const lastPriorityTools = new Set([
		'images-analyze-images',
		'images-generate-images',
		'obot-search'
	]);
	const aIsLastPriority = lastPriorityTools.has(a);
	const bIsLastPriority = lastPriorityTools.has(b);
	if (aIsLastPriority && !bIsLastPriority) return 1;
	if (!aIsLastPriority && bIsLastPriority) return -1;
	return 0;
};
