import { buildCreateUserInput, buildUpdateUserInput, emptyUserForm, userToForm, validateUserForm } from "./userForm";

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function runTests(): void {
  validatesRequiredEmail();
  buildsNormalizedCreateInput();
  buildsNormalizedUpdateInput();
  mapsUserToForm();
}

function validatesRequiredEmail(): void {
  const form = emptyUserForm();
  assert(validateUserForm(form) === "Email is required.", "expected missing email validation");
  assert(validateUserForm({ ...form, email: "owner" }) === "Email must be valid.", "expected invalid email validation");
  assert(validateUserForm({ ...form, email: "owner@example.com" }) === null, "expected valid email");
}

function buildsNormalizedCreateInput(): void {
  const input = buildCreateUserInput({
    email: " User@Example.com ",
    displayName: " Test User ",
  });

  assert(input.email === "user@example.com", "expected normalized create email");
  assert(input.display_name === "Test User", "expected trimmed create display name");
}

function buildsNormalizedUpdateInput(): void {
  const input = buildUpdateUserInput({
    email: " User@Example.com ",
    displayName: " Updated User ",
  });

  assert(input.email === "user@example.com", "expected normalized update email");
  assert(input.display_name === "Updated User", "expected trimmed update display name");
}

function mapsUserToForm(): void {
  const user = {
    email: "user@example.com",
    display_name: "User One",
  };
  const form = userToForm(user);

  assert(form.email === user.email, "expected user email");
  assert(form.displayName === user.display_name, "expected user display name");
}

runTests();
