// place files you want to import through the `$lib` alias in this folder.

// check if formData field is string
export function isString(data: FormDataEntryValue | null): data is string {
	return data !== null && typeof data === 'string';
}
