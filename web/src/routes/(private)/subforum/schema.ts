import { object, string, pipe, nonEmpty, maxLength, file, mimeType, maxSize } from 'valibot'

export const formSchema = object({
  name: pipe(
    string(),
    nonEmpty('Subforum name is required'),
    maxLength(255, 'Subforum name max characters is 255'),
  ),
  description: pipe(
    string(),
    nonEmpty('Subforum description is required'),
    maxLength(500, 'Subforum description max characters is 500'),
  ),
  icon: pipe(
    file('Subforum icon is required'),
    mimeType(['image/jpeg', 'image/png'], 'Please select a JPEG or PNG file.'),
    maxSize(1024 * 1024 * 2, 'Please select a file smaller than 2 MB.'),
  ),
  banner: pipe(
    file('Subforum banner is required'),
    mimeType(['image/jpeg', 'image/png'], 'Please select a JPEG or PNG file.'),
    maxSize(1024 * 1024 * 2, 'Please select a file smaller than 2 MB.'),
  ),
})
