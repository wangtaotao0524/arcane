export interface ProjectMeta {
	id: string;
	name: string;
	createdAt: string;
	updatedAt: string;
	autoUpdate?: boolean;
	dirName?: string;
	path: string;
}

export interface ProjectPort {
	PublicPort?: number;
	PrivatePort?: number;
	Type?: string;
	[key: string]: any;
}

export interface NetworkSettings {
	Networks: Record<
		string,
		{
			IPAddress?: string;
			Driver?: string;
			[key: string]: any;
		}
	>;
}

export interface ProjectService {
	container_id: string;
	name: string;
	status: string;
	image?: string;
	ports?: string[];
	networks?: string[];
	volumes?: string[];
	environment?: Record<string, string>;
	restart_count?: number;
	health?: string;
	networkSettings?: NetworkSettings;
}

export interface Project {
	id: string;
	name: string;
	path: string;
	runningCount: string;
	serviceCount: string;
	status: string;
	updatedAt: string;
	createdAt: string;
	autoUpdate?: boolean;
	services?: ProjectService[];
	composeContent?: string;
	envContent?: string;
}

// export interface Project {
// 	id: string;
// 	name: string;
// 	services?: ProjectService[];
// 	serviceCount?: number;
// 	runningCount?: number;
// 	status: 'running' | 'stopped' | 'partially running' | 'unknown';
// 	isExternal?: boolean;
// 	createdAt?: string;
// 	updatedAt?: string;
// 	composeContent?: string;
// 	envContent?: string;
// 	isLegacy?: boolean;
// 	agentId?: string;
// 	agentHostname?: string;
// 	isRemote?: boolean;
// 	dirName?: string;
// 	lastPolled?: string;
// 	path?: string;
// }
