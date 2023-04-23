import { useState } from "react";
import { Pagination, Badge, Tooltip, Empty } from "antd";
import axios from "axios";
import styles from "./styles.module.scss";

export default function Cluster() {
  const [pageData, setPageData] = useState<any>([]);
  const [searchParams, setSearchParams] = useState({
    pageSize: 10,
    page: 1,
  });
  async function getPageData() {
    const data = await axios(`/apis/cluster.karbour.com/v1beta1/clusters`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      params: {},
    });
    setPageData(data || {});
  }

  useState(() => {
    getPageData();
  });

  function handleChangePage(page: number, pageSize: number) {
    setSearchParams({
      ...searchParams,
      page,
      pageSize,
    });
  }

  function handleMore(item: any) {
    console.log(item, "====handleMore====");
  }

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        {pageData?.items?.map((item: any, index: number) => {
          return (
            <div className={styles.card} key={`${item.name}_${index}`}>
              <div className={styles.header}>
                <div className={styles.headerLeft}>
                  {item.metadata?.name}
                  <Badge
                    style={{
                      marginLeft: 20,
                      fontWeight: "normal",
                      color: item?.status?.healthy === "true" ? "green" : "red",
                    }}
                    status={
                      item?.status?.healthy === "true" ? "success" : "error"
                    }
                    text={item?.status?.healthy === "true" ? "健康" : "不健康"}
                  />
                </div>
                <div
                  className={styles.headerRight}
                  onClick={() => handleMore(item)}
                >
                  More
                </div>
              </div>
              <div className={styles.cardBody}>
                <div className={styles.item}>
                  <div className={styles.itemLabel}>Endpoint: </div>
                  <Tooltip title={item.spec?.access?.endpoint}>
                    <div className={styles.itemValue}>
                      {item.spec?.access?.endpoint}
                    </div>
                  </Tooltip>
                </div>
                <div className={styles.stat}>
                  <div className={styles.node}>
                    Nodes: {item?.status?.node || "--"}
                  </div>
                  <div className={styles.deloy}>
                    Delay: {item?.status?.delay || "--"}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
      {
        pageData?.items && pageData?.items?.length > 0 &&
        <div className={styles.footer}>
          <Pagination
            total={pageData?.items?.length}
            showTotal={(total, range) => `${range[0]}-${range[1]} 共 ${total} 条`}
            pageSize={searchParams?.pageSize}
            current={searchParams?.page}
            onChange={handleChangePage}
          />
        </div>
      }
      {
        !pageData?.items || !pageData?.items?.length && <Empty />
      }
    </div>
  );
}