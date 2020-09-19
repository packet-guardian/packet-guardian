import $ from "@/jlib2";
import "@/manage";

$("[name='delegated-accounts']").change(
    e =>
        (window.location.href = `/manage/${
            (e.target as HTMLSelectElement)?.value
        }`)
);
