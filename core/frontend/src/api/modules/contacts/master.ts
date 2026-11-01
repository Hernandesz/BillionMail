import { instance } from '@/api'

export interface CentralGroup { id: number; name: string }
export interface CentralTag { id: number; name: string }
export interface CentralContact {
  id: number
  email: string
  full_name: string
  job_title: string
  department: string
  company_name: string
  industry: string
  phone: string
  website: string
  linkedin_url: string
  city: string
  state_region: string
  country: string
  verification_status: string
  consent_status: string
  native_status: 'subscribed' | 'unconfirmed' | 'not_subscribed'
  source: 'master' | 'native' | 'both'
  updated_at: number
  groups: CentralGroup[]
  tags: CentralTag[]
  metadata: Record<string, unknown>
  source_details: unknown[]
}

export interface CentralContactParams {
  page: number
  page_size: number
  keyword: string
  filter: 'all' | 'master' | 'native' | 'ungrouped' | 'grouped' | 'unverified' | 'subscribed'
}

// A directory record does not imply marketing opt-in or email verification.
// Uses BillionMail's authenticated /api route and existing axios interceptors.
export const getCentralContacts = (params: CentralContactParams) =>
  instance.get<{ list: CentralContact[]; total: number }>('/contact/master/list', { params })
