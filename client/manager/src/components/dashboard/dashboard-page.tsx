import React, { useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IApplication } from '@polo/common/state/models';
import { Application } from './application/application';
import { values } from 'mobx';
import { useHistory } from 'react-router';
import {ApplicationSelectorModal} from "@/components/dashboard/application/selector/application-selector-modal";
import {useModal} from "@/components/modal/modal-hooks";

type TProps = {
    app: IApp;
}
export const Dashboard = observer((props: TProps) => {

    const [selectedAppIndex, setSelectedAppIndex] = useState(-1);
    const {show, hide} = useModal();

    const selectedApplicationLocalStorageKey = 'selected-application-name';

    const requestData = async () => {
        await props.app.retrieveStatusData();
    }

    const applications = Array.from(props.app.applications.values());

    const applicationSelectorModalName = 'applications-selector';

    useEffect(() => {
        const apps = applications;
        if (selectedAppIndex > -1 || apps.length === 0) return;
        const applicationName = localStorage.getItem(selectedApplicationLocalStorageKey);
        const foundIndex = apps
            .findIndex(app => app.configuration.name === applicationName);
        if (foundIndex > -1) {
            setSelectedAppIndex(foundIndex);
        } else {
            setSelectedAppIndex(0);
        }
    }, [applications.length, selectedAppIndex]);

    useEffect(() => {
        requestData();
        const interval = setInterval(() => requestData(), 2000);
        return () => clearInterval(interval);
    }, []);

    const openApplication = (name: string, index: number) => () => {
        setSelectedAppIndex(index);
        localStorage.setItem(selectedApplicationLocalStorageKey, name);
        hide();
    };

    const openApplicationSelectorModal = () => {
        show(applicationSelectorModalName);
    };

    const selected: IApplication = applications[selectedAppIndex];

    return <div className="font-quicksand w-full py-8 pb-12">
        <div className="w-full mx-auto lg:max-w-1500 px-5">
            <div className="flex">
                <div className="py-0 pr-5 hidden lg:block flex-shrink-0 w-3/12">
                <div className="mb-3 text-lg lg:text-xl font-medium text-nord1 dark:text-nord5">Applications</div>
                    {applications.map((application, index) =>
                        <a
                            key={index}
                            className={`block cursor-pointer px-5 py-3 rounded-md text-sm lg:text-base mb-3
                            ${selectedAppIndex === index ? 'bg-nord4 dark:bg-nord0' : ''}`}
                            onClick={openApplication(application.configuration.name, index)}>{application.configuration.name}</a>)}
                </div>
                {!!selected && <div className="flex-grow min-w-0">
                    <Application
                        sessions={props.app.sessionsByApplicationName[selected.configuration.name]}
                        failures={props.app.failures.byApplicationName[selected.configuration.name]}
                        application={selected}
                        moreThanOneApplication={applications.length > 1}
                        onApplicationsSelectorClick={openApplicationSelectorModal} />
                </div>}
            </div>
        </div>

        <ApplicationSelectorModal
            modalName={applicationSelectorModalName}
            applications={applications}
            onApplicationClick={(name, index) => openApplication(name, index)()}/>
    </div>;
})

export default Dashboard;