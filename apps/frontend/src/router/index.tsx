import {
  ClerkProvider,
} from "@clerk/clerk-react";
import {
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";
import AppLayout from "../layouts/AppLayout";
import DashboardLayout from "../layouts/DashboardLayout";
import AdminHomePage from "../pages/manage";
import HomePage from "../pages";
import RoleBasedGuard from "../components/clerk/RoleBasedGuard";

const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      {
        path: "/",
        element: <HomePage />,
      },
    ],
  },
  {
    path: "/manage",
    element: (
      <RoleBasedGuard role="admin">
        <DashboardLayout />
      </RoleBasedGuard>
    ),
    children: [
      {
        path: "/manage",
        element: <AdminHomePage />,
      },
    ],
  },
]);

function AppRouter() {
  return (
    <ClerkProvider publishableKey={clerkPubKey}>
      <RouterProvider router={router} />
    </ClerkProvider>
  );
}

export default AppRouter;