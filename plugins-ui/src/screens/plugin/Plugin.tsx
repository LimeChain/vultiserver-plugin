import PolicyForm from "../../modules/policy/components/policy-form/PolicyForm";
import { PolicyProvider } from "../../modules/policy/context/PolicyProvider";
import PolicyTable from "../../modules/policy/components/policy-table/PolicyTable";

const Plugin = () => {
  return (
    <PolicyProvider>
      <div className="left-section">
        <PolicyForm />
      </div>
      <div className="right-section">
        <PolicyTable />
      </div>
    </PolicyProvider>
  );
};

export default Plugin;
