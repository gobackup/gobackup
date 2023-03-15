export default (props: {
  name: string;
  mode?: 'line' | 'fill';
  className?: string;
  loading?: boolean;
}) => {
  const { mode = 'line', className = '', loading = false } = props;

  let classes = `ricon ri-${props.name}-${mode} ${className}`;
  if (loading) {
    classes += ' ricon-loading';
  }
  return <i className={classes}></i>;
};
