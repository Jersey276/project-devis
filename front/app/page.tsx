import { AppLayout } from "./layout";

export default function dashboardPage() {
  return (
  <AppLayout>
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      <p>Welcome to your dashboard!</p>
    </div>
  </AppLayout>
  );
}
