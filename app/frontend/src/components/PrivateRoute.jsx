import { Navigate } from "react-router-dom";

export default ({ isActive, children }) => {
  if (!isActive) return <Navigate to="/sign-in" replace />;

  return children;
};
