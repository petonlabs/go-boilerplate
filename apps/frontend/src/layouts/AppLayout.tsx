import { Outlet } from "react-router-dom";

function AppLayout() {
  return (
    <div>
      <header>
        {/* Add header content here */}
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}

export default AppLayout;