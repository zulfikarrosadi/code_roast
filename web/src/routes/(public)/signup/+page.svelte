<script lang="ts">
  import { Input } from "$lib/components/ui/input";
  import * as Alert from "$lib/components/ui/alert";
  import * as Card from "$lib/components/ui/card";
  import { applyAction, enhance } from "$app/forms";
  import { goto } from "$app/navigation";

  let user = $state({
    name: "",
    email: "",
    password: "",
    passwordConfirmation: "",
  });
  let passwordConfirmationCheck = $derived(
    user.password !== user.passwordConfirmation
      ? "Password and confirmation is not match"
      : null,
  );

  let { form } = $props()
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>Sign Up</Card.Title>
    <Card.Description
      >Enter your information to create an account</Card.Description
    >
  </Card.Header>
  <Card.Content class="space-y-4">
    {#if !form?.success && form?.message}
      <Alert.Root variant="destructive">
        <Alert.Title>Sign Up failed!</Alert.Title>
        <Alert.Description>
          {form?.message}
        </Alert.Description>
      </Alert.Root>
    {/if}
    <form method="post" class="space-y-4" action="?/register" use:enhance={(event) => {
      return async ({ result }) => {
        if (result.type === 'redirect') {
          goto('/', {invalidateAll: true})
        } else {
          await applyAction(result)
        }
      };
    }}>
      <div class={[form?.details?.fullname ? "text-red-500" : "", "space-y-2"]}>
        <label for="name">Name</label>
        <Input
          type="text"
          name="name"
          id="name"
          bind:value={user.name}
          placeholder="Your name"
          required
          aria-required="true"
          autofocus={!!form?.details?.fullname}
        />
        {#if form?.details?.fullname}
          <p class="text-sm">{form?.details?.fullname}</p>
        {/if}
      </div>
      <div class={[form?.details?.email ? "text-red-500" : "", "space-y-2"]}>
        <label for="email">Email</label>
        <Input
          type="email"
          name="email"
          bind:value={user.email}
          id="email"
          placeholder="example@email.com"
          required
          aria-required="true"
          autofocus={!!form?.details?.email}
          class={[
            form?.details?.email
              ? "text-red-500 outline outline-red-500"
              : "",
          ]}
        />
        {#if form?.details?.email}
          <p class="text-sm">{form?.details?.email}</p>
        {/if}
      </div>
      <div
        class={[
          form?.details?.password || passwordConfirmationCheck
            ? "text-red-500"
            : "",
          "space-y-2",
        ]}
      >
        <label for="password">Password</label>
        <Input
          type="password"
          name="password"
          id="password"
          required
          aria-required="true"
          bind:value={user.password}
          autofocus={!!form?.details?.password}
          class={[
            form?.details?.password || passwordConfirmationCheck
              ? "text-red-500 outline outline-red-500"
              : "",
          ]}
        />
        {#if form?.details?.password || passwordConfirmationCheck}
          <p class="text-sm">
            {form?.details?.password || passwordConfirmationCheck}
          </p>
        {/if}
      </div>
      <div>
        <label for="passwordConfirmation">Password Confirmation</label>
        <Input
          type="password"
          name="passwordConfirmation"
          id="passwordConfirmation"
          required
          bind:value={user.passwordConfirmation}
          aria-required="true"
        />
      </div>
      <button class="w-full">Sign Up</button>
    </form>
    <div class="relative flex py-5 items-center">
      <div class="flex-grow border-t border-gray-400"></div>
      <span class="text-sm flex-shrink mx-4 text-gray-400">Or</span>
      <div class="flex-grow border-t border-gray-400"></div>
    </div>
    <div class="mt-4 flex justify-center gap-1">
      <p>Already have an account?</p>
      <a href="/auth" class="underline">Sign In instead</a>
    </div>
  </Card.Content>
</Card.Root>
