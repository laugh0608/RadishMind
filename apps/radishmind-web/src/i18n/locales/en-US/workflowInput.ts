export const workflowInput = {
  "structuredRuntimeInput": "Structured runtime input",
  "notSet": "Not set",
  "clear": "Clear",
  "inputsAreRetainedOnlyForTheCurrentRequestDurableRecords": "Inputs are retained only for the current request. Durable records contain only the contract, field names and types, byte counts and digest.",
  "error_unknown_field": "This field is not part of the current input contract.",
  "error_required": "This field is required.",
  "error_boolean": "Choose true or false.",
  "error_string": "This field must be a string.",
  "error_integer": "This field must be an integer.",
  "error_number": "This field must be a number.",
  "error_string_budget": "Text must not exceed 4096 bytes.",
  "error_secret_material": "Inputs must not contain credentials, tokens, passwords or connection strings.",
  "error_integer_syntax": "Enter a decimal integer.",
  "error_integer_range": "The integer is outside the safe range.",
  "error_number_syntax": "Enter a finite JSON number."
} as const;
