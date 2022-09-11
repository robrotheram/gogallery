import { Badge, Space, Table, Tag } from "antd"
import axios from "axios";
import React, { useEffect, useState } from 'react';
import { config } from "../../store";
import { notify } from "../../store/actions";
const { Column, ColumnGroup } = Table;
const test = [
    {
        "name": "albums",
        "done": true,
        "start": "2022-09-10T19:47:53.842094663+01:00",
        "end": "2022-09-10T19:47:54.092477954+01:00",
        "duration": 250383285
    },
    {
        "name": "images",
        "done": false,
        "start": "2022-09-10T19:47:58.417235735+01:00",
        "end": "2022-09-10T19:47:58.423342357+01:00",
        "duration": 6106621
    },
    {
        "name": "photos",
        "done": true,
        "start": "2022-09-10T19:47:54.106178396+01:00",
        "end": "2022-09-10T19:47:58.391295965+01:00",
        "duration": 4285117574
    }
  ];

const formatDuration = ms => {
    if (ms < 0) ms = -ms;
    const time = {
      day: Math.floor(ms / 86400000),
      hour: Math.floor(ms / 3600000) % 24,
      minute: Math.floor(ms / 60000) % 60,
      second: Math.floor(ms / 1000) % 60,
      millisecond: Math.floor(ms) % 1000
    };
    return Object.entries(time)
      .filter(val => val[1] !== 0)
      .map(([key, val]) => `${val} ${key}${val !== 1 ? 's' : ''}`)
      .join(', ');
  };

export const TaskViewer = () => {
    const [data, setData] = useState([])

    const getTasks = () => {
        axios.get(config.baseUrl+"/tasks").then((resp)=>{
            console.log("Task data", resp.data)
            setData(resp.data)
        }).catch((err)=>{
            notify("warning", "Error from server: "+err)
        })
    }


    useEffect(() => {
        let interval = setInterval(() => {
            getTasks()
        }, 500);
        return () => {
            clearInterval(interval);
        };
    }, []);

    const sortTasks = (tasks) => {
        return tasks.sort((a,b) => {return  new Date(Date.parse(b.start)) - new Date(Date.parse(a.start));});
    }

    return (
        <Table dataSource={sortTasks(data)} pagination={false}>
            <Column title="Task Name" dataIndex="name" key="taskname" />
            <Column title="Status" dataIndex="done" key="complete" render={(done) => {
                return  done ? <><Badge status="success"/> Complete </> : <><Badge status="processing"/> In progress </>
            }}/>
            <Column title="Started At" dataIndex="start" key="started" render={(timeStr) => { 
                let date = new Date(Date.parse(timeStr));
                return date.toLocaleString();
                }
            } />
            <Column title="Time Taken" dataIndex="duration" key="duration" render={(timeStr) => { 
                return formatDuration(timeStr/1000000);
                }
            } />
           
        </Table>
    )
}