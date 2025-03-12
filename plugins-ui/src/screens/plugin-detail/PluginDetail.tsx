import Button from "@/modules/core/components/ui/button/Button";
import { useNavigate } from "react-router-dom";

const PluginDetail = () => {
  const navigate = useNavigate();

  return (
    <div className="only-section">
      <Button
        size="small"
        type="button"
        styleType="primary"
        onClick={() => navigate(`/plugin/1`)}
      >
        Open Plugin
      </Button>
    </div>
  );
};

export default PluginDetail;
