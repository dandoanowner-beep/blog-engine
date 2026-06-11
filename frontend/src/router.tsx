import { createBrowserRouter, Outlet } from 'react-router-dom'
import Layout from './components/Layout'
import PrivateRoute from './components/PrivateRoute'
import Feed from './pages/Feed'
import Login from './pages/Login'
import Register from './pages/Register'
import VerifyEmail from './pages/VerifyEmail'
import ResetPassword from './pages/ResetPassword'
import BlogDetail from './pages/BlogDetail'
import Editor from './pages/Editor'
import Profile from './pages/Profile'
import Search from './pages/Search'
import Admin from './pages/Admin'
import Portfolio from './pages/Portfolio'
import Author from './pages/Author'
import Categories from './pages/Categories'
import Forum from './pages/Forum'

function AppShell() {
  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      { index: true, element: <Feed /> },
      { path: 'blog/:id', element: <BlogDetail /> },
      { path: 'portfolio', element: <Portfolio /> },
      { path: 'author', element: <Author /> },
      { path: 'categories', element: <Categories /> },
      { path: 'forum', element: <Forum /> },
      { path: 'search', element: <Search /> },
      { path: 'profile/:username', element: <Profile /> },
      {
        path: 'editor',
        element: <PrivateRoute><Editor /></PrivateRoute>,
      },
      {
        path: 'editor/:id',
        element: <PrivateRoute><Editor /></PrivateRoute>,
      },
      {
        path: 'admin',
        element: <PrivateRoute roles={['admin', 'owner']}><Admin /></PrivateRoute>,
      },
    ],
  },
  { path: '/auth/login', element: <Login /> },
  { path: '/auth/register', element: <Register /> },
  { path: '/auth/verify', element: <VerifyEmail /> },
  { path: '/auth/forgot-password', element: <ResetPassword /> },
  { path: '/auth/reset-password', element: <ResetPassword /> },
])
