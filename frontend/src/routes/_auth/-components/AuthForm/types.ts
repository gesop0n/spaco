export type AuthMode = "register" | "login";
export type AuthField = "email" | "password";
export type FieldErrors = Partial<Record<AuthField, string>>;
