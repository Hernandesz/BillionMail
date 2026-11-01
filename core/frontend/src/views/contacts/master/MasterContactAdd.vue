<template>
  <n-modal v-model:show="visible" preset="card" title="Añadir contacto al directorio central" style="width: min(650px, 94vw)">
    <n-alert type="info" class="mb-16px">
      Guardar un correo en el directorio no lo verifica ni lo suscribe a campañas. No se asignará ningún grupo.
    </n-alert>
    <n-form label-placement="top">
      <n-grid :cols="2" :x-gap="12">
        <n-form-item-gi label="Correo electrónico *" :span="2">
          <n-input v-model:value="form.email" placeholder="contacto@empresa.com" />
        </n-form-item-gi>
        <n-form-item-gi label="Nombre"><n-input v-model:value="form.full_name" /></n-form-item-gi>
        <n-form-item-gi label="Empresa"><n-input v-model:value="form.company_name" /></n-form-item-gi>
        <n-form-item-gi label="Cargo"><n-input v-model:value="form.job_title" /></n-form-item-gi>
        <n-form-item-gi label="Departamento"><n-input v-model:value="form.department" /></n-form-item-gi>
        <n-form-item-gi label="Teléfono"><n-input v-model:value="form.phone" /></n-form-item-gi>
        <n-form-item-gi label="Industria"><n-input v-model:value="form.industry" /></n-form-item-gi>
        <n-form-item-gi label="Sitio web"><n-input v-model:value="form.website" /></n-form-item-gi>
        <n-form-item-gi label="LinkedIn"><n-input v-model:value="form.linkedin_url" /></n-form-item-gi>
        <n-form-item-gi label="Ciudad"><n-input v-model:value="form.city" /></n-form-item-gi>
        <n-form-item-gi label="Estado"><n-input v-model:value="form.state_region" /></n-form-item-gi>
        <n-form-item-gi label="País"><n-input v-model:value="form.country" /></n-form-item-gi>
      </n-grid>
    </n-form>
    <template #footer>
      <n-space justify="end">
        <n-button @click="visible = false">Cancelar</n-button>
        <n-button type="primary" :loading="saving" :disabled="!validEmail" @click="save">Guardar sin grupo</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { Message } from '@/utils'
import { importCentralContacts, type CentralContactImport } from '@/api/modules/contacts/master'

const visible = defineModel<boolean>('show', { default: false })
const emit = defineEmits<{ saved: [] }>()
const saving = ref(false)
const blank = (): CentralContactImport => ({
  email: '', full_name: '', company_name: '', job_title: '',
  department: '', industry: '', phone: '', website: '',
  linkedin_url: '', city: '', state_region: '', country: '',
})
const form = reactive<CentralContactImport>(blank())
const validEmail = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(form.email.trim()))

const save = async () => {
  if (!validEmail.value || saving.value) return
  saving.value = true
  try {
    await importCentralContacts([{ ...form, email: form.email.trim().toLowerCase() }])
    Object.assign(form, blank())
    visible.value = false
    emit('saved')
  } catch (error) {
    Message.error('No se pudo guardar el contacto. Revisa los datos y la conexión.')
  } finally {
    saving.value = false
  }
}
</script>
