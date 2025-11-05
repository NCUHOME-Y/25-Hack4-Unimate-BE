/**
 * 页面底部组件
 */
export const Footer: React.FC = () => {
  return (
    <footer className="bg-gray-50 border-t border-gray-200 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <p className="text-center text-sm text-gray-500">
          © {new Date().getFullYear()} My App. All rights reserved.
        </p>
      </div>
    </footer>
  );
};
