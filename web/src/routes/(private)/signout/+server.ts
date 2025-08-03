import { PUBLIC_API_BASE_URL } from '$env/static/public'
import { json, type RequestHandler } from '@sveltejs/kit'

export const DELETE: RequestHandler = async ({ fetch, cookies }) => {
  try {
    const res = await fetch(`${PUBLIC_API_BASE_URL}/signout`, {
      method: 'delete',
      credentials: 'include',
      headers: {
        Authorization: `Bearer ${cookies.get('access_token')}`,
      },
    })

    if (res.status !== 204) {
      return json({
        status: 'fail',
        error: {
          message: 'Gagal untuk logout, silahkan coba kembali nanti',
        },
      })
    }

    return json({
      status: 'success',
    })
  } catch (error: unknown) {
    console.log('signout api fail: ', error)
    return json({
      status: 'fail',
      error: {
        message: 'Terjadi kesalahan, silahkan coba kembali nanti',
      },
    })
  }
}
