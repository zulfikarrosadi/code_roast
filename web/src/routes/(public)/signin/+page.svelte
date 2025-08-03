<script lang="ts">
  import { Input } from '$lib/components/ui/input'
  import * as Alert from '$lib/components/ui/alert'
  import * as Card from '$lib/components/ui/card'
  import * as Form from '$lib/components/ui/form'
  import { applyAction, enhance } from '$app/forms'
  import { goto } from '$app/navigation'
  import { CircleAlertIcon, LoaderCircleIcon } from '@lucide/svelte'

  let email = $state('')
  let password = $state('')
  let formLoading = $state(false)

  let { form } = $props()
</script>

<div class="w-auto h-full flex items-center justify-center flex-col">
  <div class="space-y-2 p-4 w-full">
    <h1 class="text-3xl font-bold md:text-center">Code Roast</h1>
    <p class="text-base/relaxed md:text-center">
      Tempat tempa mental dan diskusi paling kereng h3h3
    </p>
  </div>
  <div class="flex flex-col md:flex-row gap-4 justify-center p-4">
    <Card.Root>
      <Card.Header>
        <Card.Title>Sign In</Card.Title>
        <Card.Description>Enter your information to sign in</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        {#if form && !form.success && form.message}
          {#if !formLoading}
            <Alert.Root variant="destructive">
              <CircleAlertIcon class="size-4" />
              <Alert.Title>Sign In failed!</Alert.Title>
              <Alert.Description>
                {form.message}
              </Alert.Description>
            </Alert.Root>
          {/if}
        {/if}

        <form
          method="post"
          class="space-y-4"
          use:enhance={(_) => {
            formLoading = true
            return async function ({ result }) {
              if (result.type == 'redirect') {
                goto(result.location || '/post', { invalidateAll: true })
              } else {
                formLoading = false
                applyAction(result)
              }
            }
          }}
        >
          <div class={[form?.details?.email ? 'text-red-500' : '']}>
            <label for="email">Email</label>
            <Input
              type="email"
              name="email"
              bind:value={email}
              id="email"
              placeholder="example@email.com"
              required
              aria-required="true"
              autofocus={!!form?.details?.email}
            />
            {#if form?.details?.email}
              <p>{form?.details?.email}</p>
            {/if}
          </div>
          <div class={[form?.details?.password ? 'text-red-500' : '']}>
            <label for="password">Password</label>
            <Input
              type="password"
              name="password"
              id="password"
              required
              aria-required="true"
              bind:value={password}
              autofocus={!!form?.details?.password}
            />
            {#if form?.details?.password}
              <p>{form?.details?.password}</p>
            {/if}
          </div>
          <Form.Button class="w-full" disabled={formLoading}>
            {#if formLoading}
              <LoaderCircleIcon class="animate-spin" />
            {/if}
            Sign In
          </Form.Button>
        </form>
        <div class="relative flex py-5 items-center">
          <div class="flex-grow border-t border-gray-400"></div>
          <span class="text-sm flex-shrink mx-4 text-gray-400">Or</span>
          <div class="flex-grow border-t border-gray-400"></div>
        </div>
        <div class="mt-4 flex justify-center gap-1">
          <p>Don't have account?</p>
          <a href="/signup" class="underline">Sign Up instead</a>
        </div>
      </Card.Content>
    </Card.Root>
    <figure class="">
      <img
        src="menggoreng.jpg"
        alt="Chef menggoreng"
        class="object-cover md:w-72 h-full rounded-lg"
        loading="lazy"
      />
      <figcaption class="mt-2 text-sm text-center text-gray-500 dark:text-gray-400">
        Mari menggoreng
      </figcaption>
    </figure>
  </div>
</div>
