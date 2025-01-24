import { ReactNode } from "react";
import "./Modal.css";
import closeIcon from "@/assets/Close.svg";

type ModalProps = {
    isOpen: boolean, children: ReactNode,
    onClose: () => void;
}

function Modal({ isOpen, children, onClose }: ModalProps) {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <button className="modal-close" onClick={onClose}>
                    <img src={closeIcon} alt="" />
                </button>
                {children}
            </div>
        </div>
    );
}

export default Modal;
