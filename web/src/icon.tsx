export default (props: {
  name: string;
  mode?: 'line' | 'fill';
  className?: string;
}) => {
  const { mode = 'line', className = '' } = props;

  let classes = `ricon ri-${props.name}-${mode} ${className}`;
  return <i className={classes}></i>;
};
