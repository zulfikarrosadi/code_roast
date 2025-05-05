import { PUBLIC_API_BASE_URL } from "$env/static/public";
import type { Response } from "$lib/schemas/api-response";
import type {
  AuthErrorDetails,
  AuthResponse,
} from "$lib/schemas/auth-response";

export const handle = async ({ event, resolve }) => {
  try {
    const response = await event.fetch(`${PUBLIC_API_BASE_URL}/users`, {
      headers: {
        Authorization: `Bearer ${event.cookies.get("access_token")}`,
      },
    });
    const user = (await response.json()) as Response<
      AuthResponse,
      AuthErrorDetails
    >;

    if (user.status === "success") {
      event.locals.user = {
        email: user.data.user.email,
        fullname: user.data.user.fullname,
        id: user.data.user.id,
      };
      return await resolve(event);
    }

    if (response.status === 401) {
      const response = await event.fetch(`${PUBLIC_API_BASE_URL}/refresh`, {
        credentials: "include",
      });
      const retryResponse = (await response.json()) as Response<
        AuthResponse,
        AuthErrorDetails
      >;
      if (retryResponse.status === "fail") {
        return await resolve(event);
      }
      event.locals.user = {
        email: retryResponse.data.user.email,
        fullname: retryResponse.data.user.fullname,
        id: retryResponse.data.user.id,
      };
    }

    return await resolve(event);
  } catch (error) {
    console.log(error, "hooks server error");
    return await resolve(event);
  }
};
