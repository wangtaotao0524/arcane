export type User = {
	id: string;
	username: string;
	passwordHash?: string;
	displayName?: string;
	email?: string;
	roles: string[];
	createdAt: string;
	lastLogin?: string;
	requirePasswordChange?: boolean;
	updatedAt?: string;
	oidcSubjectId?: string;
};

export interface UserRole {
	id: string;
	name: string;
	description?: string;
	permissions: Permission[];
}

export interface Permission {
	id: string;
	name: string;
	resource: string;
	action: string;
	description?: string;
}

export interface UserDto {
	username: string;
	email?: string;
	password: string;
	displayName?: string;
	role: UserRole;
	permissions?: string[];
}

export interface UpdateUserDto {
	email?: string;
	displayName?: string;
	role?: UserRole;
	permissions?: string[];
	isActive?: boolean;
}

export type CreateUser = Omit<
	User,
	| 'id'
	| 'createdAt'
	| 'updatedAt'
	| 'lastLogin'
	| 'oidcSubjectId'
	| 'passwordHash'
	| 'requirePasswordChange'
	| 'roles'
> & {
	password: string;
	roles?: string[];
};
