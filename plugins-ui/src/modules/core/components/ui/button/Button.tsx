import { ReactNode } from "react";
import "./Button.css";

type ButtonProps = {
    type: "primary" | "secondary" | "tertiary",
    size: "small" | "medium"
    children: ReactNode,
    onClick: () => any,
    style?: {},
}

const Button = ({ type, size, children, onClick, style }: ButtonProps) => {

    return (
        <button
            onClick={onClick}
            className={`button ${type} ${size}`}
            style={style}
        >
            {children}
        </button>
    );
};

export default Button;