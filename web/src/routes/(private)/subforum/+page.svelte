<script lang="ts">
  import CreateSubforumForm from '$lib/components/create-subforum-form.svelte'

  let { data } = $props()
  const CREATE_SUBFORUM_PERMISSION = 1

  $effect(() => {
    $inspect(data)
  })

  let hasCreateAccess = $derived.by(() => {
    if (!data.user) {
      return false
    }
    const access =
      data.user.role.filter((role) => role.id === CREATE_SUBFORUM_PERMISSION).length > 0
    return access
  })
</script>

<section class="mx-auto w-10/12 mt-5">
  <header class="flex flex-col md:flex-row md:justify-between">
    <h1 class="text-2xl font-bold">List of all subforums</h1>
    {#if hasCreateAccess}
      <CreateSubforumForm createForm={data.form} />
    {/if}
  </header>
</section>
