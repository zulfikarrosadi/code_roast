import { fail, superValidate } from 'sveltekit-superforms'
import { valibot } from 'sveltekit-superforms/adapters'
import { formSchema } from './schema'
import type { Actions, PageServerLoad } from './$types'
import { PUBLIC_API_BASE_URL } from '$env/static/public'
import { isRedirect, redirect } from '@sveltejs/kit'

export const load: PageServerLoad = async () => {
  const form = await superValidate(valibot(formSchema))

  return {
    form,
  }
}

export const actions: Actions = {
  default: async (event) => {
    const form = await superValidate(event, valibot(formSchema))
    if (!form.valid) {
      return fail(400, { form, message: 'Fail to create Subforum. Validation error' })
    }

    try {
      const formData = new FormData()
      formData.append('name', form.data.name)
      formData.append('description', form.data.description)
      formData.append('icon', form.data.icon)
      formData.append('banner', form.data.banner || null)

      const response = await event.fetch(`${PUBLIC_API_BASE_URL}/subforums`, {
        method: 'post',
        headers: {
          Authorization: `Bearer ${event.cookies.get('access_token')}`,
        },
        body: formData,
      })
      const result = await response.json()
      console.log(result)

      form.data.icon = undefined
      form.data.banner = undefined
      if (response.status === 201) {
        throw redirect(303, '/subforum')
      }

      return fail(400, { form, message: 'Fail to create subforum' })
    } catch (error) {
      if (isRedirect(error)) {
        return redirect(303, '/subforum')
      }
      console.log(error)
      return fail(400, { form, message: 'Fail to create subforum' })
    }
  },
}
