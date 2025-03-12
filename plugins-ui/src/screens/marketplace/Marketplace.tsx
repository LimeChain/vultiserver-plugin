import Button from "@/modules/core/components/ui/button/Button";
import { useNavigate } from "react-router-dom";

const Marketplace = () => {
  const navigate = useNavigate();

  return (
    <div className="only-section">
      <Button
        size="small"
        type="button"
        styleType="primary"
        onClick={() => navigate(`/plugin-detail/1`)}
      >
        Open
      </Button>
    </div>
  );
};

export default Marketplace;
