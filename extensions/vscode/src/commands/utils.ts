export const validateNotEmpty = (value: string): string | null => {
  return value.length === 0 ? 'Input cannot be empty' : null
}
