import { FC } from 'react';
import { Link } from 'react-router-dom';
import Icon from '../icon';

const PageTitle: FC<{
  title: string;
  backTo?: string;
}> = ({ title, backTo = '/' }) => {
  return (
    <div className="flex items-center space-x-3 mb-6 border-b-2 border-b-gray-200 py-3">
      <div>
        <Link
          to={backTo}
          className="text-2xl hover:text-red rounded hover:border-gray-100"
        >
          <Icon name="arrow-left" />
        </Link>
      </div>
      <div className="text-xl">{title}</div>
    </div>
  );
};

export default PageTitle;
