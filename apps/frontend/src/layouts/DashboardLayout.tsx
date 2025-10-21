import { Outlet } from "react-router-dom";
import { UserButton } from "@clerk/clerk-react";

function DashboardLayout() {
  return (
    <div>
      <header>
        <UserButton />
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}

export default DashboardLayout;