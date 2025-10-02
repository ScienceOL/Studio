import { Xyzen, useXyzen } from "@sciol/xyzen";

export default function App() {
  const { openXyzen, isXyzenOpen } = useXyzen();

  const handleOpenXyzen = () => {
    openXyzen();
  };

  return (
    <main className="h-full w-full">
      <div className="font-bold bg-black">App3</div>
      <button className="w-10 h-10" onClick={handleOpenXyzen}>Open Xyzen</button>
      <Xyzen />
    </main>
  );
}
