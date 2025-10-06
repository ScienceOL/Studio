export interface InteractivePanelProps extends BasicNodeProps {
  open: boolean;
  setOpen: (value: boolean) => void;
  idx: number;
  tabItems: {
    name: string;
    jsonSchema?: {
      schema: any;
      uiSchema: any;
    };
  }[];
}

export interface ResultsPanelProps extends BasicNodeProps {
  open: boolean;
  setOpen: (value: boolean) => void;
  idx: number;
}
