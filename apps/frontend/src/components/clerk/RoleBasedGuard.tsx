import { useUser } from "@clerk/clerk-react";
import { Navigate } from "react-router-dom";
import React from "react";

interface RoleBasedGuardProps {
  children: React.ReactNode;
  role: string;
}

function RoleBasedGuard({ children, role }: RoleBasedGuardProps) {
  const { user, isLoaded } = useUser();

  if (!isLoaded) {
    return <div>Loading...</div>;
  }

  if (!user) {
    return <Navigate to="/sign-in" />;
  }

  const userRole = user.publicMetadata?.role;

  if (userRole !== role) {
    return <Navigate to="/" />;
  }

  return <>{children}</>;
}

export default RoleBasedGuard;