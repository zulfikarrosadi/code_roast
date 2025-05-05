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

      // we need to this because sveltekit doesn't forwarding Set-Cookie header
      // cookie setting can only be done by event.cookies.set function
      event.cookies.set("refresh_token", retryResponse.data.refresh_token, {
        path: "/",
        httpOnly: true,
        secure: true,
        sameSite: "none",
        maxAge: 604800, // Match API's week in seconds
      });

      // we need this so access_token can be globally available by event.cookies
      event.cookies.set("access_token", retryResponse.data.access_token, {
        path: "/",
        maxAge: 3600,
        httpOnly: false,
        secure: true,
        sameSite: "strict",
      });

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
