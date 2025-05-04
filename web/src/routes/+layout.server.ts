import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async (event) => {
  if (event.locals && event.locals.user) {
    return {
      user: {
        fullname: event.locals.user.fullname,
        email: event.locals.user.email,
        id: event.locals.user.id
      }
    }
  }

  return {
    user: null
  }
}