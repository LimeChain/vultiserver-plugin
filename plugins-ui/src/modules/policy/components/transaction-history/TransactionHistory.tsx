import { useEffect, useState } from "react";
import { usePolicies } from "../../context/PolicyProvider";
import { PolicyTransactionHistory } from "../../models/policy";
import "./TransactionHistory.css";
import { publish } from "@/utils/eventBus";

type TransactionHistoryProps = {
  policyId: string;
};

const formatDate = (dateString: string): { date: string; time: string } => {
  const dateObj = new Date(dateString);
  return {
    date: dateObj.toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }), // "26 Feb 2025"
    time: dateObj.toLocaleTimeString("en-GB", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }), // "14:25:35"
  };
};

const TransactionHistory = ({ policyId }: TransactionHistoryProps) => {
  const { getPolicyHistory } = usePolicies();
  const [historyData, setHistoryData] = useState<PolicyTransactionHistory[]>(
    []
  );

  useEffect(() => {
    const fetchPolicyHistory = async (): Promise<void> => {
      try {
        const fetchedHistory = await getPolicyHistory(policyId);
        setHistoryData(fetchedHistory);
      } catch (error) {
        if (error instanceof Error) {
          console.error("Failed to get policy history:", error.message);
          publish("onToast", {
            message: error.message || "Failed to get policy history",
            type: "error",
          });
        }
      }
    };

    fetchPolicyHistory();
  }, []);

  return (
    <div className="history-panel">
      <h2>Transaction History</h2>
      <ul>
        {historyData &&
          historyData.map((item) => {
            const { date, time } = formatDate(item.updated_at);
            return (
              <li key={item.id} className="history-item">
                <span className="history-status">{item.status}</span>
                <span className="history-date">{date}</span>
                <span className="history-time">{time}</span>
              </li>
            );
          })}
        {!historyData && (
          <li key={1} className="history-item">
            Nothing to see here yet.
          </li>
        )}
      </ul>
    </div>
  );
};

export default TransactionHistory;
