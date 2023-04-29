import { useParams } from "react-router-dom";
import { modules } from "../SampleData";

export default function Module() {
  const { id } = useParams();
  const module = modules.find((module) => module.id === Number(id));
  return (
    <div className="py-4">
      <h2 className="text-base font-semibold dark:text-white">
        {module?.teamName}
      </h2>
    </div>
  );
}
