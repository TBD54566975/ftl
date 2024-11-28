package xyz.block.ftl.java.test.database;

import java.sql.Timestamp;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;

import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import io.quarkus.hibernate.orm.panache.PanacheEntity;

@Entity
@Table(name = "requests")
public class Request extends PanacheEntity {
    public String data;

    @Column(name = "created_at")
    @CreationTimestamp
    public Timestamp createdAt;

    @Column(name = "updated_at")
    @UpdateTimestamp
    public Timestamp updatedAt;

}
