<script lang="ts">
  import "../app.css";
  import { Toaster } from "$lib/components/ui/sonner/index.js";

  let { data, children } = $props();
</script>

<header class="flex h-12 w-full items-center border-b-2">
  <nav class="mx-auto flex w-10/12 items-center justify-between">
    <div>
      <a href="/">Code Roast</a>
    </div>
    <div class="flex gap-4">
      {#if data.user}
        <a href="/signout">Sign Out</a>
        <p class="font-bold">{data.user.fullname}</p>
      {:else}
        <a href="/signup">Sign Up</a>
        <a href="/signin">Sign In</a>
      {/if}
    </div>
  </nav>
</header>

<Toaster richColors closeButton expand={true} position="top-right" />

<main class="w-full mx-auto">
  {@render children()}
</main>
