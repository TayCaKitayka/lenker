export interface UserFormState {
  email: string;
  displayName: string;
}

export interface UserFormSource {
  email: string;
  display_name: string;
}

export interface UserFormInput {
  email: string;
  display_name: string;
}

export function emptyUserForm(): UserFormState {
  return {
    email: "",
    displayName: "",
  };
}

export function userToForm(user: UserFormSource): UserFormState {
  return {
    email: user.email,
    displayName: user.display_name,
  };
}

export function validateUserForm(form: UserFormState): string | null {
  const email = form.email.trim();

  if (!email) {
    return "Email is required.";
  }
  if (!email.includes("@")) {
    return "Email must be valid.";
  }
  return null;
}

export function buildCreateUserInput(form: UserFormState): UserFormInput {
  return {
    email: form.email.trim().toLowerCase(),
    display_name: form.displayName.trim(),
  };
}

export function buildUpdateUserInput(form: UserFormState): UserFormInput {
  return {
    email: form.email.trim().toLowerCase(),
    display_name: form.displayName.trim(),
  };
}
