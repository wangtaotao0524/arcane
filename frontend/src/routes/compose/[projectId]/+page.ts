import { environmentAPI } from '$lib/services/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		const projectResponse = await environmentAPI.getProject(params.projectId);

		const project = projectResponse;

		if (!project) {
			return {
				stack: null,
				error: 'Project not found',
				editorState: {
					name: '',
					composeContent: '',
					envContent: '',
					originalName: '',
					originalComposeContent: '',
					originalEnvContent: ''
				},
				servicePorts: {},
				agents: [],
				settings: {}
			};
		}

		// Set up editor state with the project data
		const editorState = {
			name: project.name || '',
			composeContent: project.composeContent || '',
			envContent: project.envContent || '',
			originalName: project.name || '',
			originalComposeContent: project.composeContent || '',
			originalEnvContent: project.envContent || ''
		};

		return {
			project,
			editorState,
			servicePorts: {}, // TODO: Extract from project services if needed
			agents: [], // TODO: Implement if needed
			settings: {
				baseServerUrl: 'localhost'
			},
			error: null
		};
	} catch (error) {
		console.error('Failed to load compose page:', error);
		return {
			stack: null,
			error: error instanceof Error ? error.message : 'Failed to load stack',
			editorState: {
				name: '',
				composeContent: '',
				envContent: '',
				originalName: '',
				originalComposeContent: '',
				originalEnvContent: ''
			},
			servicePorts: {},
			agents: [],
			settings: {}
		};
	}
};
