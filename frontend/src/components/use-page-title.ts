import { useEffect } from "react";

// usePageTitle sets the title of the browser tab. A screen reader reads the new title after a change of page.
export function usePageTitle(title: string) {
  useEffect(() => {
    document.title = `${title} - go-react-template`;
  }, [title]);
}
