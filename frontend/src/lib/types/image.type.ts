export interface ImageUpdateInfoDto {
	hasUpdate: boolean;
	updateType: string;
	currentVersion: string;
	latestVersion: string;
	currentDigest: string;
	latestDigest: string;
	checkTime: string; // ISO string from backend
	responseTimeMs: number;
	error: string;
}

export interface ImageSummaryDto {
	id: string;
	repoTags: string[];
	repoDigests: string[];
	created: number;
	size: number;
	virtualSize: number;
	labels: Record<string, unknown> | null;
	inUse: boolean;
	repo: string;
	tag: string;
	updateInfo?: ImageUpdateInfoDto;
}
