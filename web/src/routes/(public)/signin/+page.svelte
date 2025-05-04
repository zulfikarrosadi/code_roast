<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import * as Alert from '$lib/components/ui/alert';
	import * as Card from '$lib/components/ui/card';
  import { applyAction, enhance } from "$app/forms";
  import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');

  let {form} = $props()
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Sign In</Card.Title>
		<Card.Description>Enter your information to sign in</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		{#if form?.success && form?.message}
			<Alert.Root variant="destructive">
				<Alert.Title>Sign In failed!</Alert.Title>
				<Alert.Description>
					{form?.message}
				</Alert.Description>
			</Alert.Root>
		{/if}

		<form method="post" class="space-y-4" use:enhance={(event) => {
      return async function({ result }) {
        if (result.type == 'redirect') {
          goto('/', {invalidateAll :true})
        } else {
          applyAction(result)
        }
      }
    }}>
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
			<button class="w-full">Sign In</button>
		</form>
    <div class="relative flex py-5 items-center">
      <div class="flex-grow border-t border-gray-400"></div>
      <span class="text-sm flex-shrink mx-4 text-gray-400">Or</span>
      <div class="flex-grow border-t border-gray-400"></div>
    </div>
		<div class="mt-4 flex justify-center gap-1">
			<p>Don't have account?</p>
			<a href="/auth" class="underline">Sign In instead</a>
		</div>
	</Card.Content>
</Card.Root>
