<script lang="ts">
  import '../app.css'
  import { Toaster } from '$lib/components/ui/sonner/index.js'
  import { toast } from 'svelte-sonner'
  import { goto } from '$app/navigation'
  import { BellIcon } from '@lucide/svelte'

  let { data, children } = $props()

  $effect(() => {
    $inspect(data)
  })

  async function handleSignOut(event: Event) {
    event.preventDefault()
    try {
      const res = await fetch('/signout', {
        method: 'delete',
        credentials: 'include',
      })
      const result = await res.json()
      if (result.status === 'fail') {
        console.log(result)
        toast.error(result.error.message)
        return
      }
      toast.success('Berhasil logout')
      goto('/', { invalidateAll: true })
      return
    } catch (error: unknown) {
      console.log('signout failed: ', error)
      toast.error('Terjadi kesalahan, silahkan coba beberapa saat lagi')
    }
  }
</script>

<header class="flex h-18 w-full items-center border-b-2">
  <nav class="mx-auto flex w-10/12 items-center justify-between">
    <div>
      <a href="/">Code Roast</a>
    </div>
    <div class="flex gap-4">
      <a href="/post">Post</a>
      <a href="/event">Event</a>
      <a href="/subforum">Subforum</a>
      {#if !data.user}
        <a href="/signup">Sign Up</a>
        <a href="/signin">Sign In</a>
      {/if}
    </div>

    {#if data.user}
      <div class="flex gap-4 items-center">
        <BellIcon size={16} aria-label="notification" />
        <p class="font-bold">{data.user.fullname}</p>
        <form action="/signout" onsubmit={(event) => handleSignOut(event)}>
          <button class="cursor-pointer">Sign Out</button>
        </form>
      </div>
    {/if}
  </nav>
</header>

<Toaster richColors closeButton expand={true} position="top-right" />

<main class="w-full mx-auto">
  {@render children()}
</main>
