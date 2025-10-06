import LogoLoading from '@/components/basic/loading';

export default function UiTestPage() {
  return (
    <div className="bg-black h-screen flex items-center justify-center">
      <LogoLoading variant="large" animationType="galaxy" />
    </div>
  );
}
