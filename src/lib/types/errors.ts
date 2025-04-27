/* The `ServiceError` class in TypeScript extends the built-in `Error` class and sets its name to
'ServiceError'. */
export class ServiceError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'ServiceError';
		// Ensure the prototype chain is correct
		Object.setPrototypeOf(this, ServiceError.prototype);
	}
}

/* The NotFoundError class is a subclass of ServiceError that represents a resource not found error. */
export class NotFoundError extends ServiceError {
	constructor(message: string = 'Resource not found') {
		super(message);
		this.name = 'NotFoundError';
		Object.setPrototypeOf(this, NotFoundError.prototype);
	}
}
