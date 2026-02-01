# Projects Frontend

Add UI for creating/managing projects and associating them with channels.

## Components

### 1. Projects page (`src/pages/ProjectsPage.tsx`)

- List all projects with name, path, description
- Create new project form (name, path, description)
- Edit/delete existing projects
- Follow the pattern of PersonaPage

### 2. Channel project association

- In the channel settings or members panel, add a section to manage project associations
- Show associated projects in channel header or sidebar
- Add/remove project associations via dropdown/select

### 3. API hooks (`src/hooks/useProjects.ts`)

- `useProjects()` — fetch and cache project list
- `useChannelProjects(channelId)` — fetch projects for a channel
- Functions: `createProject`, `updateProject`, `deleteProject`, `addChannelProject`, `removeChannelProject`

### 4. Types (`src/types.ts`)

Add `Project` and `ChannelProject` types.

### 5. Navigation

Add "Projects" link to the sidebar/nav alongside existing items.
