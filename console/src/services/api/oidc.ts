import { api } from './client'
import type { UserPermissions } from './workspace'

export interface OIDCGroupMapping {
  oidc_group: string
  role: 'owner' | 'member'
  all_workspaces: boolean
  permissions: UserPermissions
}

interface GetGroupMappingsResponse {
  mappings: OIDCGroupMapping[]
}

interface SetGroupMappingsResponse {
  message: string
}

export const DEFAULT_PERMISSIONS: UserPermissions = {
  contacts: { read: true, write: true },
  lists: { read: true, write: true },
  templates: { read: true, write: true },
  broadcasts: { read: true, write: true },
  transactional: { read: true, write: true },
  workspace: { read: true, write: true },
  message_history: { read: true, write: true },
  blog: { read: true, write: true },
  automations: { read: true, write: true },
  llm: { read: true, write: true }
}

export const oidcApi = {
  getGroupMappings: () => api.get<GetGroupMappingsResponse>('/api/oidc.getGroupMappings'),
  setGroupMappings: (mappings: OIDCGroupMapping[]) =>
    api.post<SetGroupMappingsResponse>('/api/oidc.setGroupMappings', { mappings })
}
