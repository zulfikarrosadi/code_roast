<script lang="ts">
  import * as Dialog from './ui/dialog'
  import * as Form from './ui/form'
  import * as Alert from './ui/alert'
  import { Input } from './ui/input'
  import { Textarea } from './ui/textarea'
  import { buttonVariants } from './ui/button'
  import { formSchema } from './create-subforum-schema'
  import { superForm, fileProxy } from 'sveltekit-superforms'
  import { valibotClient } from 'sveltekit-superforms/adapters'
  import { goto } from '$app/navigation'
  import { toast } from 'svelte-sonner'
  import { applyAction } from '$app/forms'
  import SuperDebug from 'sveltekit-superforms/SuperDebug.svelte'
  import { AlertCircleIcon, Loader2Icon } from '@lucide/svelte'

  let { createForm } = $props()

  let createSubforumError = $state<string | null>(null)
  let createSubforumDialogOpen = $state(false)
  let createSubforumIsSubmit = $state(false)

  const form = superForm(createForm, {
    id: 'create-subforum',
    validators: valibotClient(formSchema),
    onSubmit: (_) => {
      createSubforumIsSubmit = true
      createSubforumError = null
    },
    onResult: async ({ result }) => {
      createSubforumIsSubmit = false
      if (result.type === 'redirect') {
        createSubforumDialogOpen = false
        toast.success('Successfully create subforum')
        goto(result.location, { invalidateAll: true })
      } else if (result.type === 'failure') {
        createSubforumError = result.data?.error
        toast.error('Fail to create subforum')
      }
      await applyAction(result)
    },
  })

  const { form: formData, enhance: formEnhance } = form
  const icon = fileProxy(formData, 'icon')
  const banner = fileProxy(formData, 'banner')
</script>

<Dialog.Root bind:open={createSubforumDialogOpen}>
  <Dialog.Trigger class={buttonVariants({ variant: 'outline' })}>Buat Subforum</Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Let's create new subforum?</Dialog.Title>
    </Dialog.Header>
    <SuperDebug data={$formData} collapsible={true} />
    {#if createSubforumError}
      <Alert.Root variant="destructive">
        <AlertCircleIcon />
        <Alert.Title>Something went wrong</Alert.Title>
        <Alert.Description>
          <p>{createSubforumError}</p>
        </Alert.Description>
      </Alert.Root>
    {/if}
    <form method="post" enctype="multipart/form-data" use:formEnhance class="space-y-4">
      <Form.Field {form} name="name">
        <Form.Control>
          {#snippet children({ props })}
            <Form.Label>Name</Form.Label>
            <Input {...props} bind:value={$formData.name} />
            <Form.FieldErrors />
          {/snippet}
        </Form.Control>
      </Form.Field>

      <Form.Field {form} name="description">
        <Form.Control>
          {#snippet children({ props })}
            <Form.Label>Description</Form.Label>
            <Textarea {...props} bind:value={$formData.description} />
            <Form.FieldErrors />
          {/snippet}
        </Form.Control>
      </Form.Field>

      <Form.Field {form} name="icon">
        <Form.Control>
          {#snippet children({ props })}
            <Form.Label>Subforum Icon</Form.Label>
            <Input {...props} type="file" bind:files={$icon} />
            <Form.FieldErrors />
          {/snippet}
        </Form.Control>
      </Form.Field>

      <Form.Field {form} name="banner">
        <Form.Control>
          {#snippet children({ props })}
            <Form.Label>Subforum Banner</Form.Label>
            <Input {...props} type="file" bind:files={$banner} />
            <Form.FieldErrors />
          {/snippet}
        </Form.Control>
      </Form.Field>
      <Form.Button disabled={createSubforumIsSubmit}>
        {#if createSubforumIsSubmit}
          Processing <div class="animate-spin"><Loader2Icon /></div>
        {:else}
          Create
        {/if}
      </Form.Button>
    </form>
  </Dialog.Content>
</Dialog.Root>
