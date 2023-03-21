import { FC } from 'react';
import { Link } from 'react-router-dom';
import Icon from '../icon';

const PageTitle: FC<{
  title: string | JSX.Element;
  backTo?: string;
  extra?: JSX.Element;
}> = ({ title, backTo = '/', extra }) => {
  return (
    <div className="flex items-center space-x-3 pb-3 justify-between">
      <div className="flex items-center gap-3">
        <Link
          to={backTo}
          className="text-2xl hover:text-red rounded hover:border-gray-100"
        >
          <Icon name="arrow-left" />
        </Link>

        <div className="text-xl">{title}</div>
      </div>

      {extra && <div>{extra}</div>}
    </div>
  );
};

export default PageTitle;
