import { useEffect, useState } from 'react'
import {
  BrowserRouter,
  Navigate,
  NavLink,
  Outlet,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  Alert,
  App as AntApp,
  Badge,
  Button,
  Card,
  Col,
  Drawer,
  Form,
  Input,
  Layout,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Statistic,
  Steps,
  Table,
  Tag,
  Typography,
  Upload,
  type TablePaginationConfig,
  type UploadFile,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { AxiosError } from 'axios'
import {
  createNotification,
  createTask,
  deleteTask,
  getDashboard,
  getNotification,
  getNotificationStats,
  getTask,
  listTasks,
  listUsers,
  login,
  markTaskInactive,
  register,
  updateTask,
  API_URL,
  GATEWAY_NAME,
  ZIPKIN_URL,
} from './api'
import { useAuthStore } from './store'
import type { AuthUser, Notification, NotificationPayload, Task, TaskPayload } from './types'

const { Header, Sider, Content } = Layout
const { Title, Text } = Typography
const queryClient = new QueryClient()

function App() {
  const hydrate = useAuthStore((state) => state.hydrate)

  useEffect(() => {
    hydrate()
  }, [hydrate])

  return (
    <QueryClientProvider client={queryClient}>
      <AntApp>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route element={<ProtectedLayout />}>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/tasks" element={<TasksPage />} />
              <Route path="/notifications" element={<NotificationsPage />} />
              <Route path="/users" element={<UsersPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AntApp>
    </QueryClientProvider>
  )
}

function ProtectedLayout() {
  const token = useAuthStore((state) => state.accessToken)
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const location = useLocation()
  const navigate = useNavigate()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  const selectedKey = location.pathname.split('/')[1] || 'dashboard'

  return (
    <Layout className="app-shell">
      <Sider className="app-sider" width={250} breakpoint="lg" collapsedWidth={0}>
        <div className="brand">
          <div className="brand-mark">LG</div>
          <div>
            <Text strong>Learn Go</Text>
            <Text type="secondary">Microservices</Text>
          </div>
        </div>
        <nav className="nav-stack">
          <NavLink className={selectedKey === 'dashboard' ? 'active' : ''} to="/dashboard">
            Dashboard
          </NavLink>
          <NavLink className={selectedKey === 'tasks' ? 'active' : ''} to="/tasks">
            Tasks
          </NavLink>
          <NavLink className={selectedKey === 'notifications' ? 'active' : ''} to="/notifications">
            Notifications
          </NavLink>
          <NavLink className={selectedKey === 'users' ? 'active' : ''} to="/users">
            Users
          </NavLink>
        </nav>
      </Sider>
      <Layout>
        <Header className="app-header">
          <div className="header-meta">
            <div>
              <Text type="secondary">Tenant</Text>
              <div className="tenant-name">{user?.tenant_slug ?? 'unknown'}</div>
            </div>
            <GatewayBadge />
          </div>
          <Space>
            <Tag color={user?.role === 'superadmin' ? 'purple' : 'blue'}>
              {user?.role ?? 'user'}
            </Tag>
            <Button
              onClick={() => {
                logout()
                navigate('/login')
              }}
            >
              Sign out
            </Button>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

function LoginPage() {
  const { message } = AntApp.useApp()
  const navigate = useNavigate()
  const setSession = useAuthStore((state) => state.setSession)
  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setSession(data)
      message.success('Login successful')
      navigate('/dashboard')
    },
    onError: (error) => message.error(errorMessage(error)),
  })

  return (
    <AuthFrame title="Sign in" subtitle="Use your Auth service account">
      <Form
        layout="vertical"
        initialValues={{ tenant_slug: 'demo' }}
        onFinish={(values) => mutation.mutate(values)}
      >
        <Form.Item name="tenant_slug" label="Tenant slug" rules={[{ required: true }]}>
          <Input placeholder="acme" />
        </Form.Item>
        <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}>
          <Input placeholder="admin@example.com" />
        </Form.Item>
        <Form.Item name="password" label="Password" rules={[{ required: true }]}>
          <Input.Password placeholder="password" />
        </Form.Item>
        <Button block type="primary" htmlType="submit" loading={mutation.isPending}>
          Sign in
        </Button>
        <Button block type="link" onClick={() => navigate('/register')}>
          Create tenant account
        </Button>
      </Form>
    </AuthFrame>
  )
}

function RegisterPage() {
  const { message } = AntApp.useApp()
  const navigate = useNavigate()
  const setSession = useAuthStore((state) => state.setSession)
  const mutation = useMutation({
    mutationFn: register,
    onSuccess: (data) => {
      setSession(data)
      message.success('Registered and signed in')
      navigate('/dashboard')
    },
    onError: (error) => message.error(errorMessage(error)),
  })

  return (
    <AuthFrame title="Register" subtitle="Create a tenant and first user">
      <Alert
        className="mb-4"
        showIcon
        type="info"
        title="New registered tenant users become owners. Superadmin dashboard routes still need a superadmin account."
      />
      <Form layout="vertical" onFinish={(values) => mutation.mutate(values)}>
        <Form.Item name="tenant_name" label="Tenant name" rules={[{ required: true }]}>
          <Input placeholder="Acme Inc" />
        </Form.Item>
        <Form.Item name="tenant_slug" label="Tenant slug" rules={[{ required: true }]}>
          <Input placeholder="acme" />
        </Form.Item>
        <Form.Item name="username" label="Username" rules={[{ required: true }]}>
          <Input placeholder="mohit" />
        </Form.Item>
        <Form.Item name="email" label="Email" rules={[{ required: true, type: 'email' }]}>
          <Input placeholder="mohit@example.com" />
        </Form.Item>
        <Form.Item name="password" label="Password" rules={[{ required: true }]}>
          <Input.Password placeholder="password" />
        </Form.Item>
        <Button block type="primary" htmlType="submit" loading={mutation.isPending}>
          Register
        </Button>
        <Button block type="link" onClick={() => navigate('/login')}>
          Back to sign in
        </Button>
      </Form>
    </AuthFrame>
  )
}

function DashboardPage() {
  const query = useQuery({ queryKey: ['dashboard'], queryFn: getDashboard, retry: false })
  const data = query.data

  return (
    <PageTitle
      title="Dashboard"
      description="Superadmin overview from task GraphQL and Auth gRPC."
    >
      {query.isError && <Alert type="warning" showIcon title={errorMessage(query.error)} />}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Users" value={data?.users.total ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Active users" value={data?.users.active ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Tasks" value={data?.tasks.total ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Active tasks" value={data?.tasks.active ?? 0} />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]} className="mt-4">
        <Col xs={24} lg={12}>
          <Card title="Tasks by status" loading={query.isLoading}>
            <Space wrap>
              {(data?.tasks.by_status ?? []).map((item) => (
                <Badge key={item.status} count={item.count} color="blue">
                  <Tag>{item.status}</Tag>
                </Badge>
              ))}
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Users by role" loading={query.isLoading}>
            <Space wrap>
              {(data?.users.by_role ?? []).map((item) => (
                <Badge key={item.role} count={item.count} color="purple">
                  <Tag>{item.role}</Tag>
                </Badge>
              ))}
            </Space>
          </Card>
        </Col>
      </Row>
    </PageTitle>
  )
}

function TasksPage() {
  const { message } = AntApp.useApp()
  const queryClient = useQueryClient()
  const currentUser = useAuthStore((state) => state.user)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<string>()
  const [drawerTask, setDrawerTask] = useState<Task | null>(null)
  const [modalTask, setModalTask] = useState<Task | null | undefined>(undefined)

  const query = useQuery({
    queryKey: ['tasks', page, search, status],
    queryFn: () =>
      listTasks({
        page,
        per_page: 10,
        search,
        status,
        only_mine: false,
      }),
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['tasks'] })
  const inactiveMutation = useMutation({
    mutationFn: markTaskInactive,
    onSuccess: () => {
      message.success('Task marked inactive')
      invalidate()
    },
    onError: (error) => message.error(errorMessage(error)),
  })
  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      message.success('Task deleted')
      invalidate()
    },
    onError: (error) => message.error(errorMessage(error)),
  })
  const taskDetailMutation = useMutation({
    mutationFn: getTask,
    onError: (error) => message.error(errorMessage(error)),
  })

  const openTaskDrawer = (id: string) => {
    taskDetailMutation.mutate(id, {
      onSuccess: (task) => setDrawerTask(task),
    })
  }

  const openTaskEditor = (id: string) => {
    taskDetailMutation.mutate(id, {
      onSuccess: (task) => setModalTask(task),
    })
  }

  const columns: ColumnsType<Task> = [
    {
      title: 'Title',
      dataIndex: 'title',
      render: (_, record) => (
        <Button type="link" className="table-link" onClick={() => openTaskDrawer(record.id)}>
          {record.title}
        </Button>
      ),
    },
    { title: 'Status', dataIndex: 'status', render: (value) => <Tag>{value}</Tag> },
    { title: 'Priority', dataIndex: 'priority', render: priorityTag },
    {
      title: 'Active',
      dataIndex: 'is_active',
      render: (value) => <Badge status={value ? 'success' : 'default'} text={value ? 'Yes' : 'No'} />,
    },
    { title: 'Updated', dataIndex: 'updated_at', render: formatDate },
    {
      title: 'Actions',
      width: 230,
      render: (_, record) => {
        const isOwner = record.user_id === currentUser?.id
        return (
          <Space>
            <Button
              size="small"
              disabled={!isOwner}
              loading={taskDetailMutation.isPending}
              onClick={() => openTaskEditor(record.id)}
            >
              Edit
            </Button>
            <Button size="small" disabled={!isOwner} onClick={() => inactiveMutation.mutate(record.id)}>
              Inactive
            </Button>
            <Popconfirm title="Delete task permanently?" onConfirm={() => deleteMutation.mutate(record.id)}>
              <Button size="small" danger disabled={!isOwner}>
                Delete
              </Button>
            </Popconfirm>
          </Space>
        )
      },
    },
  ]

  return (
    <PageTitle title="Tasks" description="Create, update, list, inactivate, and delete tasks.">
      <Card>
        <div className="toolbar">
          <Space wrap>
            <Input.Search
              allowClear
              placeholder="Search tasks"
              onSearch={(value) => {
                setPage(1)
                setSearch(value)
              }}
              style={{ width: 260 }}
            />
            <Select
              allowClear
              placeholder="Status"
              value={status}
              onChange={(value) => {
                setPage(1)
                setStatus(value)
              }}
              style={{ width: 160 }}
              options={[
                { value: 'todo', label: 'Todo' },
                { value: 'in_progress', label: 'In progress' },
                { value: 'done', label: 'Done' },
              ]}
            />
          </Space>
          <Button type="primary" onClick={() => setModalTask(null)}>
            New task
          </Button>
        </div>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={query.data?.items ?? []}
          loading={query.isLoading}
          pagination={{
            current: query.data?.meta.page ?? page,
            pageSize: query.data?.meta.per_page ?? 10,
            total: query.data?.meta.total ?? 0,
            showSizeChanger: false,
          }}
          onChange={(pagination: TablePaginationConfig) => setPage(pagination.current ?? 1)}
          scroll={{ x: 860 }}
        />
      </Card>
      <TaskModal
        task={modalTask}
        open={modalTask !== undefined}
        onClose={() => setModalTask(undefined)}
        onSaved={() => {
          setModalTask(undefined)
          invalidate()
        }}
      />
      <Drawer title={drawerTask?.title} open={!!drawerTask} onClose={() => setDrawerTask(null)}>
        {drawerTask && (
          <Space direction="vertical" size="middle" className="full-width">
            <Text>{drawerTask.description || 'No description'}</Text>
            <Space>
              <Tag>{drawerTask.status}</Tag>
              {priorityTag(drawerTask.priority)}
            </Space>
            {drawerTask.image_url && <img className="task-image" src={drawerTask.image_url} alt="" />}
            <Text type="secondary">Created {formatDate(drawerTask.created_at)}</Text>
            <Text type="secondary">Updated {formatDate(drawerTask.updated_at)}</Text>
          </Space>
        )}
      </Drawer>
    </PageTitle>
  )
}

function TaskModal({
  task,
  open,
  onClose,
  onSaved,
}: {
  task: Task | null | undefined
  open: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const { message } = AntApp.useApp()
  const [form] = Form.useForm()
  const [selectedFiles, setSelectedFiles] = useState<UploadFile[]>([])
  const saveMutation = useMutation({
    mutationFn: (payload: TaskPayload) => (task ? updateTask(task.id, payload) : createTask(payload)),
    onSuccess: () => {
      message.success(task ? 'Task updated' : 'Task created')
      form.resetFields()
      setSelectedFiles([])
      onSaved()
    },
    onError: (error) => message.error(errorMessage(error)),
  })

  useEffect(() => {
    if (!open) return
    form.setFieldsValue({
      title: task?.title ?? '',
      description: task?.description ?? '',
      status: task?.status ?? 'todo',
      priority: task?.priority ?? 'normal',
    })
  }, [form, open, task])

  const handleClose = () => {
    form.resetFields()
    setSelectedFiles([])
    onClose()
  }

  return (
    <Modal
      title={task ? 'Edit task' : 'New task'}
      open={open}
      onCancel={handleClose}
      okText={task ? 'Save' : 'Create'}
      confirmLoading={saveMutation.isPending}
      onOk={() => form.submit()}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => {
          saveMutation.mutate({
            ...values,
            image: selectedFiles[0]?.originFileObj,
          })
        }}
      >
        <Form.Item name="title" label="Title" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="description" label="Description">
          <Input.TextArea rows={4} />
        </Form.Item>
        <Row gutter={12}>
          <Col span={12}>
            <Form.Item name="status" label="Status">
              <Select
                options={[
                  { value: 'todo', label: 'Todo' },
                  { value: 'in_progress', label: 'In progress' },
                  { value: 'done', label: 'Done' },
                ]}
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="priority" label="Priority">
              <Select
                options={[
                  { value: 'low', label: 'Low' },
                  { value: 'normal', label: 'Normal' },
                  { value: 'high', label: 'High' },
                ]}
              />
            </Form.Item>
          </Col>
        </Row>
        {task?.image_url && (
          <div className="task-image-preview">
            <Text type="secondary">Current image</Text>
            <img className="task-image" src={task.image_url} alt={task.title} />
          </div>
        )}
        <Upload
          listType="picture"
          beforeUpload={() => false}
          maxCount={1}
          fileList={selectedFiles}
          onChange={({ fileList }) => setSelectedFiles(fileList)}
        >
          <Button>{task?.image_url ? 'Replace image' : 'Select image'}</Button>
        </Upload>
      </Form>
    </Modal>
  )
}

function UsersPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [role, setRole] = useState<string>()
  const query = useQuery({
    queryKey: ['users', page, search, role],
    queryFn: () => listUsers({ page, per_page: 10, search, role }),
    retry: false,
  })

  const columns: ColumnsType<AuthUser> = [
    { title: 'Username', dataIndex: 'username' },
    { title: 'Email', dataIndex: 'email' },
    { title: 'Tenant', dataIndex: 'tenant_slug' },
    { title: 'Role', dataIndex: 'role', render: (value) => <Tag>{value}</Tag> },
    {
      title: 'Active',
      dataIndex: 'is_active',
      render: (value) => <Badge status={value ? 'success' : 'default'} text={value ? 'Yes' : 'No'} />,
    },
    { title: 'Created', dataIndex: 'created_at', render: formatDate },
  ]

  return (
    <PageTitle title="Users" description="Superadmin user list from Auth service.">
      {query.isError && <Alert type="warning" showIcon title={errorMessage(query.error)} />}
      <Card>
        <div className="toolbar">
          <Space wrap>
            <Input.Search
              allowClear
              placeholder="Search users"
              onSearch={(value) => {
                setPage(1)
                setSearch(value)
              }}
              style={{ width: 260 }}
            />
            <Select
              allowClear
              placeholder="Role"
              value={role}
              onChange={(value) => {
                setPage(1)
                setRole(value)
              }}
              style={{ width: 160 }}
              options={[
                { value: 'superadmin', label: 'Superadmin' },
                { value: 'admin', label: 'Admin' },
                { value: 'owner', label: 'Owner' },
                { value: 'staff', label: 'Staff' },
                { value: 'guest', label: 'Guest' },
              ]}
            />
          </Space>
        </div>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={query.data?.items ?? []}
          loading={query.isLoading}
          pagination={{
            current: query.data?.meta.page ?? page,
            pageSize: query.data?.meta.per_page ?? 10,
            total: query.data?.meta.total ?? 0,
            showSizeChanger: false,
          }}
          onChange={(pagination: TablePaginationConfig) => setPage(pagination.current ?? 1)}
          scroll={{ x: 820 }}
        />
      </Card>
    </PageTitle>
  )
}

function NotificationsPage() {
  const { message } = AntApp.useApp()
  const queryClient = useQueryClient()
  const [lastNotification, setLastNotification] = useState<Notification | null>(null)

  const statsQuery = useQuery({
    queryKey: ['notification-stats'],
    queryFn: getNotificationStats,
    refetchInterval: 1500,
  })

  const statusQuery = useQuery({
    queryKey: ['notification', lastNotification?.id],
    queryFn: () => getNotification(lastNotification!.id),
    enabled: Boolean(lastNotification?.id),
    refetchInterval: (query) =>
      query.state.data?.status === 'delivered' || query.state.data?.status === 'failed'
        ? false
        : 1000,
  })

  const createMutation = useMutation({
    mutationFn: createNotification,
    onSuccess: (data) => {
      setLastNotification(data)
      message.success('Notification queued')
      queryClient.invalidateQueries({ queryKey: ['notification-stats'] })
    },
    onError: (error) => message.error(errorMessage(error)),
  })

  const current = statusQuery.data ?? lastNotification
  const stats = statsQuery.data

  return (
    <PageTitle
      title="Notifications"
      description="Small worker-pool service for learning goroutines, channels, and background jobs."
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card title="Queue notification">
            <Form
              layout="vertical"
              initialValues={{ recipient: 'dev@example.com', message: 'Hello from a channel job' }}
              onFinish={(values: NotificationPayload) => createMutation.mutate(values)}
            >
              <Form.Item name="recipient" label="Recipient" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="message" label="Message" rules={[{ required: true }]}>
                <Input.TextArea rows={4} />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
                Send to worker pool
              </Button>
            </Form>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card title="Worker stats" loading={statsQuery.isLoading}>
            <Row gutter={[12, 12]}>
              <Col xs={12} md={8}>
                <Statistic title="Workers" value={stats?.workers ?? 0} />
              </Col>
              <Col xs={12} md={8}>
                <Statistic title="Queue size" value={stats?.queue_size ?? 0} />
              </Col>
              <Col xs={12} md={8}>
                <Statistic title="Waiting" value={stats?.jobs_waiting ?? 0} />
              </Col>
              <Col xs={12} md={8}>
                <Statistic title="Queued" value={stats?.queued ?? 0} />
              </Col>
              <Col xs={12} md={8}>
                <Statistic title="Sending" value={stats?.sending ?? 0} />
              </Col>
              <Col xs={12} md={8}>
                <Statistic title="Delivered" value={stats?.delivered ?? 0} />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      <Card title="Last notification">
        {current ? (
          <Space direction="vertical" size="middle" className="full-width">
            <Steps
              current={notificationStep(current.status)}
              status={current.status === 'failed' ? 'error' : 'process'}
              items={[{ title: 'Queued' }, { title: 'Sending' }, { title: 'Delivered' }]}
            />
            <Space wrap>
              <Tag>{current.status}</Tag>
              {current.worker_id ? <Tag color="blue">worker {current.worker_id}</Tag> : null}
              <Text type="secondary">{current.id}</Text>
            </Space>
            <Text>{current.message}</Text>
            {current.error ? <Alert type="error" showIcon title={current.error} /> : null}
          </Space>
        ) : (
          <Text type="secondary">No notification submitted yet.</Text>
        )}
      </Card>
    </PageTitle>
  )
}

function AuthFrame({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle: string
  children: React.ReactNode
}) {
  return (
    <div className="auth-page">
      <Card className="auth-card">
        <Title level={2}>{title}</Title>
        <Text type="secondary">{subtitle}</Text>
        <div className="mt-4">
          <GatewayBadge />
        </div>
        <div className="mt-6">{children}</div>
      </Card>
    </div>
  )
}

function PageTitle({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <Space direction="vertical" size="large" className="full-width">
      <div>
        <Title level={2} className="page-title">
          {title}
        </Title>
        <Text type="secondary">{description}</Text>
      </div>
      {children}
    </Space>
  )
}

function GatewayBadge() {
  return (
    <Space size={6} wrap className="gateway-badge">
      <Tag color={GATEWAY_NAME === 'Kong' ? 'cyan' : 'default'}>{GATEWAY_NAME}</Tag>
      <Text type="secondary">{API_URL}</Text>
      <Text type="secondary">Zipkin {ZIPKIN_URL}</Text>
    </Space>
  )
}

function priorityTag(priority: string) {
  const color = priority === 'high' ? 'red' : priority === 'low' ? 'green' : 'blue'
  return <Tag color={color}>{priority}</Tag>
}

function notificationStep(status: string) {
  if (status === 'queued') return 0
  if (status === 'sending') return 1
  return 2
}

function formatDate(value: string) {
  if (!value) return '-'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function errorMessage(error: unknown) {
  const axiosError = error as AxiosError<{ message?: string; error?: unknown }>
  if (axiosError.response?.data?.message) {
    return axiosError.response.data.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Something went wrong'
}

export default App
