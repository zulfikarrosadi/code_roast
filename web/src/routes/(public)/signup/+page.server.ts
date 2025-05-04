import type { Actions } from "./$types";
import { PUBLIC_API_BASE_URL } from "$env/static/public";
import { fail, isRedirect, redirect } from "@sveltejs/kit";
import type { Response } from "$lib/response-schema";

type AuthErrorDetails = {
  email: string;
  password: string;
  fullname: string;
};
type AuthResponse = {
  user: {
    id: string;
    fullname: string;
    email: string;
  };
  access_token: string;
  refresh_token: string;
};

export const actions: Actions = {
  register: async ({ request, fetch, cookies, locals }) => {
    const data = await request.formData();
    const fullname = data.get("name");
    const email = data.get("email");
    const password = data.get("password");
    const password_confirmation = data.get("password_confirmation");

    try {
      const res = await fetch(`${PUBLIC_API_BASE_URL}/api/v1/signup`, {
        method: "post",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          fullname: fullname?.toString(),
          email: email?.toString(),
          password: password?.toString(),
          password_confirmation: password_confirmation?.toString(),
        }),
      });
      const result = (await res.json()) as Response<
        AuthResponse,
        AuthErrorDetails
      >;
      if (result.status === "fail") {
        return fail(res.status, {
        success:false,
          message: result.error.message,
          details: {
            email: result.error.details?.email,
            fullname: result.error.details?.fullname,
            password: result.error.details?.password,
          },
        });
      }
      // we need to this because sveltekit doesn't forwarding Set-Cookie header
      // cookie setting can only be done by event.cookies.set function
      cookies.set("refresh_token", result.data.refresh_token, {
        path: "/",
        httpOnly: true,
        secure: true,
        sameSite: "none",
        maxAge: 604800, // Match API's week in seconds
      });

      // we need this so access_token can be globally available by event.cookies
      cookies.set("access_token", result.data.access_token, {
        path: "/",
        maxAge: 3600,
        httpOnly: false,
        secure: true,
        sameSite: "strict",
      });
      return redirect(303, '/')
    } catch (error) {
      if (isRedirect(error)) {
        throw error;
      }
      fail(500, {
        success:false,
        message: "something went wrong, please try again later",
      });
    }
  },
};
