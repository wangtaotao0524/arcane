export const load = async ({ parent }) => {
	const { settings } = await parent();

	return {
		settings
	};
};
