import { useState, useEffect, useCallback } from 'react'
import {
  Descriptions,
  Button,
  Table,
  Input,
  Select,
  Switch,
  Space,
  App,
  Collapse,
  Popconfirm,
  Spin,
  Typography
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlusOutlined,
  DeleteOutlined
} from '@ant-design/icons'
import { useLingui } from '@lingui/react/macro'
import { SettingsSectionHeader } from './SettingsSectionHeader'
import { oidcApi, DEFAULT_PERMISSIONS } from '../../services/api/oidc'
import type { OIDCGroupMapping } from '../../services/api/oidc'
import type { UserPermissions } from '../../services/api/workspace'

const { Text } = Typography

function PermissionsEditor({
  permissions,
  onChange
}: {
  permissions: UserPermissions
  onChange: (p: UserPermissions) => void
}) {
  const { t } = useLingui()

  const update = (resource: string, type: 'read' | 'write', value: boolean) => {
    onChange({
      ...permissions,
      [resource]: {
        ...(permissions as unknown as Record<string, { read: boolean; write: boolean }>)[resource],
        [type]: value
      }
    })
  }

  const data = Object.entries(permissions).map(([resource, perms]) => ({
    key: resource,
    resource: resource.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
    read: perms.read,
    write: perms.write
  }))

  return (
    <Table
      dataSource={data}
      pagination={false}
      size="small"
      className="border border-gray-200 rounded-md"
      columns={[
        { title: t`Resource`, dataIndex: 'resource', key: 'resource', width: '40%' },
        {
          title: t`Read`,
          dataIndex: 'read',
          key: 'read',
          width: '30%',
          render: (value: boolean, record: { key: string }) => (
            <Switch
              checked={value}
              onChange={(checked) => update(record.key, 'read', checked)}
              size="small"
            />
          )
        },
        {
          title: t`Write`,
          dataIndex: 'write',
          key: 'write',
          width: '30%',
          render: (value: boolean, record: { key: string }) => (
            <Switch
              checked={value}
              onChange={(checked) => update(record.key, 'write', checked)}
              size="small"
            />
          )
        }
      ]}
    />
  )
}

export function OIDCSettings() {
  const { t } = useLingui()
  const { message } = App.useApp()

  const [mappings, setMappings] = useState<OIDCGroupMapping[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  const fetchMappings = useCallback(async () => {
    setLoading(true)
    try {
      const res = await oidcApi.getGroupMappings()
      setMappings(res.mappings)
      setDirty(false)
    } catch {
      message.error(t`Failed to load group mappings`)
    } finally {
      setLoading(false)
    }
  }, [message, t])

  useEffect(() => {
    if (window.OIDC_ENABLED) {
      fetchMappings()
    }
  }, [fetchMappings])

  const addMapping = () => {
    setMappings((prev) => [
      ...prev,
      {
        oidc_group: '',
        role: 'member',
        all_workspaces: true,
        permissions: { ...DEFAULT_PERMISSIONS }
      }
    ])
    setDirty(true)
  }

  const removeMapping = (index: number) => {
    setMappings((prev) => prev.filter((_, i) => i !== index))
    setDirty(true)
  }

  const updateMapping = (index: number, patch: Partial<OIDCGroupMapping>) => {
    setMappings((prev) => prev.map((m, i) => (i === index ? { ...m, ...patch } : m)))
    setDirty(true)
  }

  const saveMappings = async () => {
    const invalid = mappings.find((m) => !m.oidc_group.trim())
    if (invalid) {
      message.warning(t`All mappings must have an IdP group name`)
      return
    }

    setSaving(true)
    try {
      await oidcApi.setGroupMappings(mappings)
      message.success(t`Group mappings saved`)
      setDirty(false)
    } catch {
      message.error(t`Failed to save group mappings`)
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <SettingsSectionHeader
        title={t`Single Sign-On (OIDC)`}
        description={t`OpenID Connect SSO authentication configured via environment variables`}
      />

      {window.OIDC_ENABLED ? (
        <>
          <Descriptions
            bordered
            column={1}
            size="small"
            styles={{ label: { width: '200px', fontWeight: '500' } }}
          >
            <Descriptions.Item label={t`OIDC SSO`}>
              <span style={{ color: '#52c41a' }}>
                <CheckCircleOutlined style={{ marginRight: '8px' }} />
                {t`Enabled`}
              </span>
            </Descriptions.Item>

            <Descriptions.Item label={t`Magic Code Login`}>
              {window.OIDC_ALLOW_MAGIC_CODE ? (
                <span style={{ color: '#52c41a' }}>
                  <CheckCircleOutlined style={{ marginRight: '8px' }} />
                  {t`Allowed`}
                </span>
              ) : (
                <span style={{ color: '#ff4d4f' }}>
                  <CloseCircleOutlined style={{ marginRight: '8px' }} />
                  {t`Disabled (SSO enforced)`}
                </span>
              )}
            </Descriptions.Item>
          </Descriptions>

          <div style={{ marginTop: 32 }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 16
              }}
            >
              <div>
                <Text strong style={{ fontSize: 16 }}>
                  {t`Group Mappings`}
                </Text>
                <br />
                <Text type="secondary" style={{ fontSize: 13 }}>
                  {t`Map IdP groups to Broadside roles and permissions. Users without a matching group get member access with full permissions by default.`}
                </Text>
              </div>
              <Space>
                {dirty && (
                  <Button type="primary" onClick={saveMappings} loading={saving}>
                    {t`Save`}
                  </Button>
                )}
                <Button icon={<PlusOutlined />} onClick={addMapping}>
                  {t`Add Mapping`}
                </Button>
              </Space>
            </div>

            {loading ? (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Spin />
              </div>
            ) : mappings.length === 0 ? (
              <div
                style={{
                  textAlign: 'center',
                  padding: '32px 16px',
                  color: '#8c8c8c',
                  border: '1px dashed #d9d9d9',
                  borderRadius: 8
                }}
              >
                <Text type="secondary">
                  {t`No group mappings configured. New SSO users will receive member access with full permissions to all workspaces.`}
                </Text>
              </div>
            ) : (
              <Collapse
                accordion
                items={mappings.map((mapping, index) => ({
                  key: String(index),
                  label: (
                    <Space>
                      <Text strong>{mapping.oidc_group || t`(unnamed group)`}</Text>
                      <Text type="secondary">→</Text>
                      <Text type="secondary">
                        {mapping.role === 'owner' ? t`Owner` : t`Member`}
                      </Text>
                    </Space>
                  ),
                  extra: (
                    <Popconfirm
                      title={t`Remove this mapping?`}
                      onConfirm={(e) => {
                        e?.stopPropagation()
                        removeMapping(index)
                      }}
                      onCancel={(e) => e?.stopPropagation()}
                    >
                      <Button
                        type="text"
                        danger
                        size="small"
                        icon={<DeleteOutlined />}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Popconfirm>
                  ),
                  children: (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
                        <div style={{ flex: '1 1 200px' }}>
                          <Text
                            type="secondary"
                            style={{ display: 'block', marginBottom: 4, fontSize: 13 }}
                          >
                            {t`IdP Group Name`}
                          </Text>
                          <Input
                            value={mapping.oidc_group}
                            placeholder={t`e.g. broadside-admins`}
                            onChange={(e) =>
                              updateMapping(index, { oidc_group: e.target.value })
                            }
                          />
                        </div>
                        <div style={{ flex: '0 0 150px' }}>
                          <Text
                            type="secondary"
                            style={{ display: 'block', marginBottom: 4, fontSize: 13 }}
                          >
                            {t`Role`}
                          </Text>
                          <Select
                            value={mapping.role}
                            onChange={(value) => updateMapping(index, { role: value })}
                            style={{ width: '100%' }}
                            options={[
                              { label: t`Owner`, value: 'owner' },
                              { label: t`Member`, value: 'member' }
                            ]}
                          />
                        </div>
                        <div style={{ flex: '0 0 180px', display: 'flex', alignItems: 'end' }}>
                          <Space>
                            <Switch
                              checked={mapping.all_workspaces}
                              onChange={(checked) =>
                                updateMapping(index, { all_workspaces: checked })
                              }
                              size="small"
                            />
                            <Text style={{ fontSize: 13 }}>{t`All workspaces`}</Text>
                          </Space>
                        </div>
                      </div>

                      <div>
                        <Text
                          type="secondary"
                          style={{ display: 'block', marginBottom: 8, fontSize: 13 }}
                        >
                          {t`Permissions`}
                        </Text>
                        <PermissionsEditor
                          permissions={mapping.permissions}
                          onChange={(permissions) => updateMapping(index, { permissions })}
                        />
                      </div>
                    </div>
                  )
                }))}
              />
            )}
          </div>
        </>
      ) : (
        <div style={{ color: '#8c8c8c', fontStyle: 'italic' }}>
          {t`OIDC single sign-on is not configured.`}{' '}
          <a
            href="https://docs.notifuse.com/installation#oidc-sso-configuration"
            target="_blank"
            rel="noopener noreferrer"
          >
            {t`Learn how to enable OIDC SSO`}
          </a>
        </div>
      )}
    </>
  )
}
