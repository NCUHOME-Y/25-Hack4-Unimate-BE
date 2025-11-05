interface HeaderProps {
  title?: string;
}

/**
 * 页面头部组件
 */
export const Header: React.FC<HeaderProps> = ({ title = 'My App' }) => {
  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <h1 className="text-xl font-semibold text-gray-900">{title}</h1>
          <nav className="flex space-x-4">
            {/* 添加导航链接 */}
          </nav>
        </div>
      </div>
    </header>
  );
};
