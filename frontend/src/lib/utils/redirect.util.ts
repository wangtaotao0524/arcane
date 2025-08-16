import type { User } from '$lib/types/user.type';
import type { Settings } from '$lib/types/settings.type';

export function getAuthRedirectPath(path: string, user: User | null, settings: Settings | null) {
	const isSignedIn = !!user;

	const isUnauthenticatedOnlyPath =
		path === '/auth/login' ||
		path.startsWith('/auth/login/') ||
		path === '/auth/oidc/login' ||
		path.startsWith('/auth/oidc/login') ||
		path === '/auth/oidc/callback' ||
		path.startsWith('/auth/oidc/callback') ||
		path === '/img' ||
		path.startsWith('/img') ||
		path === '/favicon.ico';

	// Root should not render; redirect explicitly
	if (path === '/') {
		return isSignedIn ? '/dashboard' : '/auth/login';
	}

	const isOnboardingPath = path === '/onboarding' || path.startsWith('/onboarding');

	if (!isSignedIn && !isUnauthenticatedOnlyPath) {
		return '/auth/login';
	}

	if (isSignedIn && settings && !settings.onboardingCompleted) {
		if (!isOnboardingPath) {
			return '/onboarding/welcome';
		}
	}

	if (isUnauthenticatedOnlyPath && isSignedIn && settings?.onboardingCompleted) {
		return '/dashboard';
	}

	if (path === '/' && isSignedIn) {
		return '/dashboard';
	}

	return null;
}

// Lightweight session validator
async function validateSession(): Promise<boolean> {
	if (typeof window === 'undefined') return true; // SSR: skip
	try {
		const res = await fetch('/api/auth/validate', { credentials: 'include' });
		return res.ok;
	} catch {
		return false;
	}
}

// Session-aware redirect
export async function getAuthRedirectPathWithSessionCheck(
	path: string,
	user: User | null,
	settings: Settings | null
): Promise<string | null> {
	let effectiveUser = user;

	if (effectiveUser) {
		const valid = await validateSession();
		if (!valid) {
			effectiveUser = null;
			// Use normal logic for signed-out state; do not return '/'
			return getAuthRedirectPath(path, null, settings);
		}
	}

	return getAuthRedirectPath(path, effectiveUser, settings);
}
