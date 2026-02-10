import { Navigate } from "react-router-dom";

export default ({ isActive, children }) => {
  if (!isActive) return <Navigate to="/" replace />;

  return children;
};
